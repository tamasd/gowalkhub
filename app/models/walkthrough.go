package models

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/coopernurse/gorp"
	"github.com/dropbox/godropbox/container/lrucache"
	"github.com/revel/revel"
)

var (
	walkthroughCache     *lrucache.LRUCache
	walkthroughCacheLock sync.RWMutex

	walkthroughMainPage     []string
	walkthroughMainPageLock sync.RWMutex
)

func init() {
	revel.OnAppStart(func() {
		walkthroughCache = lrucache.New(revel.Config.IntDefault("walkthroughcache.size", 262144))
		walkthroughMainPage = []string{}
	})
}

func ResetWalkthroughMainPage(txn gorp.SqlExecutor) {
	walkthroughMainPageLock.Lock()
	defer walkthroughMainPageLock.Unlock()

	walkthroughMainPage = []string{}

	res, err := txn.Select(Walkthrough{}, "SELECT UUID FROM Walkthrough ORDER BY Created DESC LIMIT 25")
	if err != nil {
		revel.ERROR.Print(err)
		return
	}

	for _, r := range res {
		wtres := r.(*Walkthrough)
		walkthroughMainPage = append(walkthroughMainPage, wtres.UUID)
	}
}

func getWalkthroughMainPage() []string {
	walkthroughMainPageLock.RLock()
	defer walkthroughMainPageLock.RUnlock()

	return walkthroughMainPage[:]
}

func putWalkthroughToCache(s Walkthrough) {
	walkthroughCacheLock.Lock()
	defer walkthroughCacheLock.Unlock()

	walkthroughCache.Set(s.UUID, s)
}

func getWalkthroughFromCache(id string) Walkthrough {
	walkthroughCacheLock.RLock()
	defer walkthroughCacheLock.RUnlock()

	wt, found := walkthroughCache.Get(id)
	if !found {
		return Walkthrough{}
	}

	walkthrough, ok := wt.(Walkthrough)
	if !ok {
		revel.WARN.Printf("Invalid data found in the walkthrough cache at key: %s\n", id)
		return Walkthrough{}
	}

	return walkthrough
}

func GetMainPageWalkthroughs(txn gorp.SqlExecutor) []Walkthrough {
	wts := []Walkthrough{}

	for _, uuid := range getWalkthroughMainPage() {
		wts = append(wts, GetWalkthrough(txn, uuid))
	}

	return wts
}

type Walkthrough struct {
	UUID        string            `json:"uuid"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]string `db:"-" json:"parameters"`
	Steps       []string          `db:"-" json:"steps"`
	Url         string            `db:"-" json:"url"`
	Author      string            `json:"author"`
	Created     int64             `json:"-"`

	RawParameters string `db:"Parameters" json:"-"`
	RawSteps      string `db:"Steps" json:"-"`
}

func CreateNewWalkthrough() *Walkthrough {
	return &Walkthrough{
		UUID:       uuid.NewRandom().String(),
		Parameters: make(map[string]string),
		Steps:      make([]string, 0),
	}
}

func (w Walkthrough) IsValid() bool {
	return w.UUID != ""
}

func (w *Walkthrough) unpack() {
	w.Steps = strings.Split(w.RawSteps, " ")
	json.Unmarshal([]byte(w.RawParameters), &(w.Parameters))
	w.Url = "walkthrough/" + w.UUID
}

func (w *Walkthrough) pack() {
	w.RawSteps = strings.Join(w.Steps, " ")
	j, _ := json.Marshal(w.Parameters)
	w.RawParameters = string(j)
}

func (w *Walkthrough) Insert(txn *gorp.Transaction) error {
	w.pack()
	w.Created = time.Now().Unix()
	putWalkthroughToCache(*w)

	if err := txn.Insert(w); err != nil {
		revel.WARN.Print(err)
		return err
	}

	ResetWalkthroughMainPage(txn)

	return nil
}

func (w *Walkthrough) Save(txn *gorp.Transaction) error {
	w.pack()
	putWalkthroughToCache(*w)

	if _, err := txn.Update(w); err != nil {
		revel.WARN.Print(err)
		return err
	}

	return nil
}

func GetWalkthrough(txn gorp.SqlExecutor, uuid string) Walkthrough {
	wt := getWalkthroughFromCache(uuid)
	if wt.IsValid() {
		revel.TRACE.Printf("Walkthrough cache HIT: %s\n", uuid)
		return wt
	} else {
		revel.TRACE.Printf("Walkthrough cache MISS: %s\n", uuid)
	}

	wtdata, err := txn.Get(Walkthrough{}, uuid)
	if err != nil {
		revel.ERROR.Print(err)
		return Walkthrough{}
	}

	if wtdata == nil {
		revel.TRACE.Printf("Walkthrough not found: %s\n", uuid)
		return Walkthrough{}
	}

	w := wtdata.(*Walkthrough)
	w.unpack()
	putWalkthroughToCache(*w)
	return *w
}
