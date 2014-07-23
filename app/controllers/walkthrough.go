package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/revel/revel"
	"github.com/tamasd/gowalkhub/app/models"
)

type Walkthrough struct {
	GorpController
}

func (c Walkthrough) GetWalkthrough(uuid string) revel.Result {
	wt := models.GetWalkthrough(c.Txn, uuid)
	if !wt.IsValid() {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}
	return c.RenderJson(wt)
}

func (c Walkthrough) SaveWalkthrough(uuid string) revel.Result {
	wt := models.GetWalkthrough(c.Txn, uuid)

	if !wt.IsValid() {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}

	user := connected(c.GorpController.Controller)
	if uid := c.Session["uid"]; uid == "" || (uid != wt.Author && user.Role != models.UserRoleAdmin) {
		c.Response.Status = http.StatusForbidden
		return c.RenderJson(nil)
	}

	requestWalkthrough := &models.Walkthrough{}
	if err := json.NewDecoder(c.Request.Body).Decode(requestWalkthrough); err != nil || requestWalkthrough.UUID != uuid || requestWalkthrough.Author != wt.Author {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(nil)
	}

	if err := requestWalkthrough.Save(c.Txn); err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(nil)
	}

	return c.RenderJson(requestWalkthrough)
}

func (c Walkthrough) ListWalkthroughs() revel.Result {
	return c.RenderJson(models.GetMainPageWalkthroughs(c.Txn))
}

func (c Walkthrough) SaveRecordedWalkthrough() revel.Result {
	author := connected(c.GorpController.Controller)
	if !author.IsValid() {
		c.Response.Status = http.StatusForbidden
		return c.RenderJson(nil)
	}

	requestData := &RecordedWalkthrough{}
	if err := json.NewDecoder(c.Request.Body).Decode(requestData); err != nil {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(nil)
	}

	wt := models.CreateNewWalkthrough()
	wt.Name = requestData.Title
	wt.Author = author.Id
	if requestData.PasswordParameter {
		wt.Parameters["password"] = ""
	}

	for i, step := range requestData.Steps {
		if i == 0 && step.CMD == "open" {
			u, err := url.Parse(step.Arg0)
			if err != nil {
				c.Response.Status = http.StatusBadRequest
				return c.RenderJson(nil)
			}
			wt.Parameters["domain"] = u.Host
			u.Host = "[domain]"
			step.Arg0 = u.String()
		}

		s := models.CreateNewStep(step.CMD, step.Arg0, step.Arg1, author.Id)
		if err := s.Insert(c.Txn); err != nil {
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJson(nil)
		}
		wt.Steps = append(wt.Steps, s.UUID)
	}

	if err := wt.Insert(c.Txn); err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(nil)
	}

	return c.RenderJson(map[string]string{"uuid": wt.UUID})
}

func (c Walkthrough) GetStep(uuid string) revel.Result {
	s := models.GetStep(c.Txn, uuid)
	if !s.IsValid() {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}
	return c.RenderJson(s)
}

func (c Walkthrough) SaveStep(uuid string) revel.Result {
	s := models.GetStep(c.Txn, uuid)

	if !s.IsValid() {
		c.Response.Status = http.StatusNotFound
		return c.RenderJson(nil)
	}

	user := connected(c.GorpController.Controller)
	if uid := c.Session["uid"]; uid == "" || (uid != s.Author && user.Role != models.UserRoleAdmin) {
		c.Response.Status = http.StatusForbidden
		return c.RenderJson(nil)
	}

	requestStep := &models.Step{}
	if err := json.NewDecoder(c.Request.Body).Decode(requestStep); err != nil || requestStep.UUID != uuid {
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(nil)
	}

	requestStep.Author = s.Author

	if err := requestStep.Save(c.Txn); err != nil {
		c.Response.Status = http.StatusInternalServerError
		revel.ERROR.Print(err)
		return c.RenderJson(nil)
	}

	return c.RenderJson(requestStep)
}

type RecordedWalkthrough struct {
	Title string
	Steps []struct {
		CMD  string
		Arg0 string
		Arg1 string
	}
	PasswordParameter bool
}
