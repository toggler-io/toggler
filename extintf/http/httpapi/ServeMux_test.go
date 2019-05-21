package httpapi_test

import (
	"github.com/adamluzsi/FeatureFlags/extintf/http/httpapi"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
)

func NewServeMux(t *testcase.T) *httpapi.ServeMux {
	return httpapi.NewServeMux(usecases.NewUseCases(GetStorage(t)))
}
