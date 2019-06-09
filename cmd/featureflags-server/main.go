package main

import (
	"fmt"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"log"
	"net/http"
)

func main() {
	storage := testing.NewTestStorage()
	useCases := usecases.NewUseCases(storage)
	mux := httpintf.NewServeMux(useCases)

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

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		log.Fatal(err)
	}
}
