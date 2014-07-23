package models

import (
	"sync"

	"code.google.com/p/go-uuid/uuid"

	"github.com/coopernurse/gorp"
	"github.com/dropbox/godropbox/container/lrucache"
	"github.com/revel/revel"
)

var (
	stepCache     *lrucache.LRUCache
	stepCacheLock sync.RWMutex
)

func init() {
	revel.OnAppStart(func() {
		stepCache = lrucache.New(revel.Config.IntDefault("stepcache.size", 262144))
	})
}

func putStepToCache(s Step) {
	stepCacheLock.Lock()
	defer stepCacheLock.Unlock()

	stepCache.Set(s.UUID, s)
}

func getStepFromCache(id string) Step {
	stepCacheLock.RLock()
	defer stepCacheLock.RUnlock()

	s, found := stepCache.Get(id)
	if !found {
		return Step{}
	}

	step, ok := s.(Step)
	if !ok {
		revel.WARN.Printf("Invalid data found in the step cache at key: %s\n", id)
		return Step{}
	}

	return step
}

type Step struct {
	UUID           string `json:"uuid"`
	CanEdit        bool   `json:"canEdit" db:"-"`
	Title          string `json:"title"`
	TitleRaw       string `json:"titleRaw"`
	ShowTitle      bool   `json:"showTitle"`
	Description    string `json:"description"`
	DescriptionRaw string `json:"descriptionRaw"`
	Command        string `json:"command"`
	CommandRaw     string `json:"commandRaw"`
	PureCommand    string `json:"pureCommand"`
	PureCommandRaw string `json:"pureCommandRaw"`
	AndWait        bool   `json:"andWait"`
	Arg1           string `json:"arg1"`
	Arg2           string `json:"arg2"`
	Highlight      string `json:"highlight"`
	Author         string `json:"-"`
}

func CreateNewStep(cmd, arg0, arg1, author string) *Step {
	s := &Step{
		UUID:           uuid.NewRandom().String(),
		CommandRaw:     cmd,
		PureCommandRaw: cmd,
		AndWait:        true,
		Arg1:           arg0,
		Arg2:           arg1,
		Author:         author,
	}

	s.filter()

	return s
}

func (s Step) IsValid() bool {
	return s.UUID != ""
}

func (s *Step) Insert(txn *gorp.Transaction) error {
	s.filter()
	putStepToCache(*s)

	if err := txn.Insert(s); err != nil {
		revel.WARN.Print(err)
		return err
	}

	return nil
}

func (s *Step) Save(txn *gorp.Transaction) error {
	s.filter()
	putStepToCache(*s)

	if _, err := txn.Update(s); err != nil {
		revel.WARN.Print(err)
		return err
	}

	return nil
}

func (s *Step) filter() {
	s.Title = plainText.Sanitize(s.TitleRaw)
	s.Description = filteredHTML.Sanitize(s.DescriptionRaw)
	s.Command = plainText.Sanitize(s.CommandRaw)
	s.PureCommand = s.Command
	s.PureCommandRaw = s.CommandRaw

	hl := ""
	if i, ok := highlightData[s.Command]; ok {
		if i {
			hl = s.Arg2
		} else {
			hl = s.Arg1
		}
	}

	s.Highlight = hl
}

func GetStep(txn gorp.SqlExecutor, id string) Step {
	step := getStepFromCache(id)
	if step.IsValid() {
		revel.TRACE.Printf("Step cache HIT: %s\n", id)
		return step
	} else {
		revel.TRACE.Printf("Step cache MISS: %s\n", id)
	}

	stepdata, err := txn.Get(Step{}, id)
	if err != nil {
		revel.ERROR.Print(err)
		return Step{}
	}

	if stepdata == nil {
		revel.TRACE.Printf("Step not found: %s\n", id)
		return Step{}
	}

	s := stepdata.(*Step)
	putStepToCache(*s)
	return *s
}

// These selenium commands have locator arguments.
// false means that the locator is the first argument, true means that the second
var highlightData = map[string]bool{
	"addSelection":        false,
	"assignId":            false,
	"check":               false,
	"click":               false,
	"clickAt":             false,
	"contextMenu":         false,
	"contextMenuAt":       false,
	"doubleClick":         false,
	"doubleClickAt":       false,
	"dragAndDrop":         false,
	"dragAndDropToObject": false,
	"dragdrop":            false,
	"fireEvent":           false,
	"focus":               false,
	"highlight":           false,
	"keyDown":             false,
	"keyPress":            false,
	"keyUp":               false,
	"mouseDown":           false,
	"mouseDownAt":         false,
	"mouseDownRight":      false,
	"mouseDownRightAt":    false,
	"mouseMove":           false,
	"mouseMoveAt":         false,
	"mouseOut":            false,
	"mouseOutAt":          false,
	"mouseOver":           false,
	"mouseUp":             false,
	"mouseUpAt":           false,
	"mouseUpRight":        false,
	"mouseUpRightAt":      false,
	"removeAllSelections": false,
	"removeSelection":     false,
	"sendKeys":            false,
	"select":              false,
	"selectFrame":         false,
	"setCursorPosition":   false,
	"submit":              false,
	"type":                false,
	"typeKeys":            false,
	"uncheck":             false,
}
