package main

import (
	"github.com/adamluzsi/FeatureFlags/extintf/httpintf"
	"github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"log"
	"net/http"
)

func main() {
	storage := testing.NewTestStorage()
	useCases := usecases.NewUseCases(storage)
	mux := httpintf.NewServeMux(useCases)

	if err := http.ListenAndServe(`:8080`, mux); err != nil {
		log.Fatal(err)
	}
}
