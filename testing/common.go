package testing

import (
	"context"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
)

const ContextLetVar = `common context of testing`
const ExampleIDLetVar = `example unique ID`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(ContextLetVar, func(t *testcase.T) interface{} {
			return context.Background()
		})

		s.Let(ExampleIDLetVar, func(t *testcase.T) interface{} {
			return fixtures.Random.String()
		})
	})
}

func GetContext(t *testcase.T) context.Context {
	return t.I(ContextLetVar).(context.Context)
}

func ExampleID(t *testcase.T) string {
	return t.I(ExampleIDLetVar).(string)
}