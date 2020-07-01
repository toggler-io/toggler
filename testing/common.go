package testing

import (
	"context"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
)

const LetVarContext = `common testing context`
const LetVarExampleID = `example unique ID`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(LetVarContext, func(t *testcase.T) interface{} {
			return context.Background()
		})

		s.Let(LetVarExampleID, func(t *testcase.T) interface{} {
			return fixtures.Random.String()
		})
	})
}

func GetContext(t *testcase.T) context.Context {
	return t.I(LetVarContext).(context.Context)
}

func ExampleID(t *testcase.T) string {
	return t.I(LetVarExampleID).(string)
}
