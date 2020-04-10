package testing

import (
	"context"

	"github.com/adamluzsi/testcase"
)

const ContextLetVar = `common context of testing`

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(ContextLetVar, func(t *testcase.T) interface{} {
			return context.Background()
		})
	})
}

func GetContext(t *testcase.T) context.Context {
	return t.I(ContextLetVar).(context.Context)
}

