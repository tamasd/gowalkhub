package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"github.com/revel/revel"
	"github.com/tamasd/gowalkhub/app/models"
)

var GOOGLEOAUTH = (*oauth.Config)(nil)

func init() {
	revel.OnAppStart(func() {
		GOOGLEOAUTH = &oauth.Config{
			ClientId:     revel.Config.StringDefault("oauth.client.id", ""),
			ClientSecret: revel.Config.StringDefault("oauth.client.secret", ""),
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://accounts.google.com/o/oauth2/token",
			RedirectURL:  revel.Config.StringDefault("oauth.redirect.url", ""),
			Scope:        "https://www.googleapis.com/auth/plus.profile.emails.read",
		}
	})
}

type User struct {
	GorpController
}

func (c User) Me() revel.Result {
	user := connected(c.GorpController.Controller)
	if !user.IsValid() {
		c.Response.Status = http.StatusForbidden
		return c.RenderJson(nil)
	}
	return c.RenderJson(user)
}

func (c User) Get(id string) revel.Result {
	user := models.GetUser(c.Txn, id)
	if !user.IsValid() {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}
	return c.RenderJson(user)
}

func (c User) Update(id string) revel.Result {
	user := connected(c.GorpController.Controller)
	if !user.IsValid() || (user.Id != id && user.Role != models.UserRoleAdmin) {
		c.Response.Status = http.StatusForbidden
		return c.RenderJson(nil)
	}

	updateUser := models.GetUser(c.Txn, id)
	if !user.IsValid() {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}

	requestUser := models.User{}
	if err := json.NewDecoder(c.Request.Body).Decode(&requestUser); err != nil || id != requestUser.Id {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(nil)
	}

	if user.Role != models.UserRoleAdmin && requestUser.Role != updateUser.Role {
		c.Response.Status = http.StatusForbidden
		return c.RenderJson(nil)
	}

	if err := requestUser.Save(c.Txn); err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(nil)
	}

	return c.RenderJson(requestUser)
}

func (c User) Login() revel.Result {
	state := c.Request.URL.Query().Get("destination")
	return c.Redirect(GOOGLEOAUTH.AuthCodeURL(state))
}

func (c User) Logout() revel.Result {
	delete(c.Session, "uid")
	return c.Redirect("/")
}

func (c User) Callback() (r revel.Result) {
	destination := c.Request.URL.Query().Get("state")
	if destination == "" {
		destination = "/"
	}
	r = c.Redirect(destination)

	code := c.Request.URL.Query().Get("code")
	t := &oauth.Transport{Config: GOOGLEOAUTH}
	_, err := t.Exchange(code)
	if err != nil {
		revel.ERROR.Println(err)
		return
	}

	client := t.Client()
	resp, err := client.Get("https://www.googleapis.com/plus/v1/people/me?fields=emails&key=" + GOOGLEOAUTH.ClientId)
	if err != nil {
		revel.ERROR.Print(err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		revel.ERROR.Print("Error retrieving oauth2 user data (%d):\n%s", resp.StatusCode, string(body))
		return
	}

	or := &Oauth2Response{}
	if err := json.Unmarshal(body, or); err != nil {
		revel.ERROR.Print(err)
		return
	}

	if len(or.Emails) < 1 {
		revel.WARN.Println("No email address found.")
		return
	}

	email := or.Emails[0].Value

	user := models.GetUserFromEmail(c.Txn, email)
	if !user.IsValid() {
		user = models.RegisterUser(c.Txn, email)
	}

	c.Session["uid"] = user.Id

	return
}

func (c User) Token() revel.Result {
	return c.RenderText(c.Session["csrf_token"])
}

func connected(c *revel.Controller) models.User {
	return c.RenderArgs["user"].(models.User)
}

func setuser(c *revel.Controller) revel.Result {
	var user models.User

	if uid, ok := c.Session["uid"]; ok {
		user = models.GetUser(DBM, uid)
	}

	c.RenderArgs["user"] = user

	return nil
}

type Oauth2Response struct {
	Emails []Oauth2ResponseEmail
}

type Oauth2ResponseEmail struct {
	Value string
	Type  string
}
