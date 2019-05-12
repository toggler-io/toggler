package testing

import (
	"github.com/adamluzsi/frameless/resources"
	_ "github.com/adamluzsi/frameless/resources/specs"

)

func NewFixtureFactory() *FixtureFactory {
	return &FixtureFactory{}
}

type FixtureFactory struct {
	resources.GenericFixtureFactory
}

func (ff *FixtureFactory) Create(EntityType interface{}) interface{} {
	switch EntityType.(type) {

	default:
		return ff.GenericFixtureFactory.Create(EntityType)
	}
}

