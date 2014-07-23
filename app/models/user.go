package models

import (
	"sync"

	"code.google.com/p/go-uuid/uuid"

	"github.com/coopernurse/gorp"
	"github.com/dropbox/godropbox/container/lrucache"
	"github.com/revel/revel"
)

var (
	cache     *lrucache.LRUCache
	cacheLock sync.RWMutex
)

func init() {
	revel.OnAppStart(func() {
		cache = lrucache.New(revel.Config.IntDefault("usercache.size", 65536))
	})
}

const (
	UserRoleDisabled uint8 = iota
	UserRoleNormal
	UserRoleAdmin
)

type User struct {
	Id    string
	Name  string
	Email string
	Role  uint8
}

func (u User) IsValid() bool {
	return u.Id != ""
}

func RegisterUser(txn *gorp.Transaction, email string) User {
	u := User{}
	u.Email = email
	u.Role = UserRoleNormal
	u.Id = uuid.NewRandom().String()

	if err := txn.Insert(&u); err != nil {
		revel.ERROR.Print(err)
		return User{}
	}

	putUserToCache(u)
	return u
}

func (u User) Save(txn *gorp.Transaction) error {
	putUserToCache(u)

	if _, err := txn.Update(&u); err != nil {
		revel.WARN.Print(err)
		return err
	}

	return nil
}

func putUserToCache(u User) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	cache.Set(u.Id, u)
}

func getUserFromCache(id string) User {
	cacheLock.RLock()
	defer cacheLock.RUnlock()

	u, found := cache.Get(id)
	if !found {
		return User{}
	}

	user, ok := u.(User)
	if !ok {
		revel.WARN.Printf("Invalid data found in the user cache at key: %s\n", id)
		return User{}
	}

	return user
}

func GetUser(txn gorp.SqlExecutor, id string) User {
	user := getUserFromCache(id)
	if user.IsValid() {
		revel.TRACE.Printf("User cache HIT: %s\n", id)
		return user
	} else {
		revel.TRACE.Printf("User cache MISS: %s\n", id)
	}

	userdata, err := txn.Get(User{}, id)
	if err != nil {
		revel.ERROR.Print(err)
		return User{}
	}

	if userdata == nil {
		revel.TRACE.Printf("User not found: %s\n", id)
		return User{}
	}

	usr := userdata.(*User)
	putUserToCache(*usr)
	return *usr
}

func GetUserFromEmail(txn gorp.SqlExecutor, email string) User {
	userdata := &User{}
	err := txn.SelectOne(userdata, "SELECT * FROM User Where Email = ?", email)
	if err != nil {
		revel.ERROR.Print(err)
		return User{}
	}

	if userdata == nil {
		revel.TRACE.Printf("User not found: %s\n", email)
		return User{}
	}

	putUserToCache(*userdata)
	return *userdata
}
