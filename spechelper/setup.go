package spechelper

import (
	"github.com/adamluzsi/testcase"
	"testing"
)

var setups []func(s *testcase.Spec)

func NewSpec(tb testing.TB) *testcase.Spec {
	s := testcase.NewSpec(tb)
	SetUp(s)
	return s
}

func SetUp(s *testcase.Spec) {
	for _, setup := range setups {
		setup(s)
	}
	Storage.Let(s, nil)
}
