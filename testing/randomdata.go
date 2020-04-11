package testing

import (
	"sync"

	"github.com/Pallinder/go-randomdata"
)

//nolint:gochecknoglobals
var mutex sync.Mutex

func RandomName() string {
	mutex.Lock()
	defer mutex.Unlock()
	return randomdata.SillyName()
}

func RandomExternalPilotID() string {
	mutex.Lock()
	defer mutex.Unlock()
	return randomdata.MacAddress()
}

func RandomUniqUserID() string {
	mutex.Lock()
	defer mutex.Unlock()
	return randomdata.MacAddress()
}
