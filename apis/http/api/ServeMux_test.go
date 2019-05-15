package api_test

import (
	"github.com/adamluzsi/FeatureFlags/apis/http/api"
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
)

func NewServeMux(t *testcase.T) *api.ServeMux {
	return api.NewServeMux(usecases.NewUseCases(GetStorage(t)))
}
