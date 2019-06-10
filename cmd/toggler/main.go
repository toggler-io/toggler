package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adamluzsi/toggler/extintf/httpintf"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/adamluzsi/toggler/testing"
	"github.com/adamluzsi/toggler/usecases"
	"github.com/unrolled/logger"
)

func main() {
	storage := testing.NewTestStorage()
	useCases := usecases.NewUseCases(storage)
	mux := httpintf.NewServeMux(useCases)

	loggerMW := logger.New()
	app := loggerMW.Handler(mux)

	i := security.Issuer{Storage: storage}
	t, err := i.CreateNewToken(`testing`, nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(t.Token)

	pu, err := useCases.ProtectedUsecases(t.Token)

	if err != nil {
		panic(err)
	}

	ff := rollouts.FeatureFlag{Name: `test`}

	if err := pu.CreateFeatureFlag(&ff); err != nil {
		panic(err)
	}

	if err := pu.SetPilotEnrollmentForFeature(ff.ID, `test-public-pilot-id-1`, true); err != nil {
		panic(err)
	}

	if err := pu.SetPilotEnrollmentForFeature(ff.ID, `test-public-pilot-id-2`, false); err != nil {
		panic(err)
	}

	if err := http.ListenAndServe(`:8080`, app); err != nil {
		log.Fatal(err)
	}
}
