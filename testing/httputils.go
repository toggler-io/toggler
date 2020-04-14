package testing

import (
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/httpspec"
)

func GivenHTTPRequestHasAppToken(s *testcase.Spec) {
	s.Before(func(t *testcase.T) {
		httpspec.Header(t).Set(`X-App-Token`, ExampleTextToken(t))
	})
}
