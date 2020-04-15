package testing

import "github.com/adamluzsi/testcase"

var setups []func(s *testcase.Spec)

func SetUp(s *testcase.Spec) {
	for _, setup := range setups {
		setup(s)
	}
}
