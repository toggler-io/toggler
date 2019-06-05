package main

import (
	"fmt"
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf"
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
		panic(err.Error())
	}
	fmt.Println(t.Token)

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		log.Fatal(err)
	}
}
