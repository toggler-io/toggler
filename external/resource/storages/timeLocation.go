package storages

import "time"

var timeLocation *time.Location

func init() {
	tl, err := time.LoadLocation(`UTC`)
	if err == nil {
		timeLocation = tl
	} else {
		timeLocation = time.Local
	}
}
