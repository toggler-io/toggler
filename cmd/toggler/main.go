package main

import (
	"fmt"
	"github.com/adamluzsi/toggler/extintf/storages"
	"log"
	"net/http"
	"os"

	"github.com/adamluzsi/toggler/extintf/httpintf"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/adamluzsi/toggler/usecases"
	"github.com/unrolled/logger"
)

func main() {
	storage, err := storages.New(connstr())
	if err != nil{
		log.Fatal(err)
	}
	defer storage.Close()

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

	pu.CreateFeatureFlag(&ff)
	pu.SetPilotEnrollmentForFeature(ff.ID, `test-public-pilot-id-1`, true)
	pu.SetPilotEnrollmentForFeature(ff.ID, `test-public-pilot-id-2`, false)

	if err := http.ListenAndServe(`:8080`, app); err != nil {
		log.Fatal(err)
	}
}

func connstr() string {
	connstr, isSet := os.LookupEnv(`DATABASE_URL`)

	if !isSet {
		log.Fatal(`please set "DATABASE_URL" to use the service`)
	}

	return connstr
}
