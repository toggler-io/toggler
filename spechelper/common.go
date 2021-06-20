package spechelper

import (
	"context"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
)

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		Context.Let(s, nil)
		ExampleID.Let(s, nil)
	})
}

var (
	Context = testcase.Var{
		Name: `common testing context`,
		Init: func(t *testcase.T) interface{} {
			return context.Background()
		},
	}
	ExampleID = testcase.Var{
		Name: `example unique ID`,
		Init: func(t *testcase.T) interface{} {
			return fixtures.Random.String()
		},
	}
)

func ContextGet(t *testcase.T) context.Context {
	return Context.Get(t).(context.Context)
}

func ExampleIDGet(t *testcase.T) string {
	return ExampleID.Get(t).(string)
}
