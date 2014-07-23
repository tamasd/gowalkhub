package models

import (
	"github.com/microcosm-cc/bluemonday"
)

var (
	filteredHTML = bluemonday.UGCPolicy()
	plainText    = bluemonday.StrictPolicy()
)
