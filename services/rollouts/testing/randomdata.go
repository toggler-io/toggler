package testing

import (
	"sync"

	"github.com/Pallinder/go-randomdata"
)

var mutex sync.Mutex

func ExampleFlagName() string {
	mutex.Lock()
	defer mutex.Unlock()
	return randomdata.SillyName()
}

func ExampleExternalPilotID() string {
	mutex.Lock()
	defer mutex.Unlock()
	return randomdata.MacAddress()
}
