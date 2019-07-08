package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/adamluzsi/toggler/extintf/caches"
	"github.com/adamluzsi/toggler/extintf/storages"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/adamluzsi/toggler/extintf/httpintf"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/adamluzsi/toggler/usecases"
	"github.com/unrolled/logger"
)

func main() {
	flagSet := flag.NewFlagSet(`toggler`, flag.ExitOnError)
	portConfValue := flagSet.String(`port`, os.Getenv(`PORT`), `set http server port else the env variable "PORT" value will be used.`)
	cmd := flagSet.String(`cmd`, `http-server`, `cli command. cmds: "http-server", "create-token".`)
	fixtures := flagSet.Bool(`create-fixtures`, false, `create default fixtures for development purpose.`)
	dbURL := flagSet.String(`database-url`, ``, `define what url should be used for the db connection. Default value used from ENV[DATABASE_URL].`)
	cacheURL := flagSet.String(`cache-url`, ``, `define what url should be used for the cache connection. default value is taken from ENV[CACHE_URL].`)

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	setupDatabaseURL(dbURL)
	setupCacheURL(cacheURL)

	storage, err := storages.New(*dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	cache, err := caches.New(*cacheURL, storage)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	if *fixtures {
		createFixtures(storage)
	}

	switch *cmd {
	case `http-server`:
		port, err := strconv.Atoi(*portConfValue)
		if err != nil {
			panic(err)
		}
		httpListenAndServe(cache, port)

	case `create-token`:
		createToken(storage, flagSet.Arg(0))

	default:
		fmt.Println(`please provide on of the commands`)
		fmt.Printf("\t%s\n", `http-server`)
		fmt.Printf("\t%s\n", `create-token`)
	}

}

func setupDatabaseURL(dbURL *string) {
	if *dbURL != `` {
		return
	}

	connstr, isSet := os.LookupEnv(`DATABASE_URL`)

	if !isSet {
		log.Fatal(`db url env variable is missing: "DATABASE_URL"`)
	}

	*dbURL = connstr
}

func setupCacheURL(cacheURL *string) {
	if *cacheURL == `` {
		*cacheURL = os.Getenv(`CACHE_URL`)
	}
}

func createFixtures(s usecases.Storage) {
	useCases := usecases.NewUseCases(s)
	issuer := security.Issuer{Storage: s}

	tt, _, err := issuer.CreateNewToken(context.Background(), `testing`, nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(tt)

	pu, err := useCases.ProtectedUsecases(context.Background(), tt)

	if err != nil {
		panic(err)
	}

	ff := rollouts.FeatureFlag{Name: `test`}
	_ = pu.CreateFeatureFlag(context.TODO(), &ff)
	_ = pu.SetPilotEnrollmentForFeature(context.Background(), ff.ID, `test-public-pilot-id-1`, true)
	_ = pu.SetPilotEnrollmentForFeature(context.Background(), ff.ID, `test-public-pilot-id-2`, false)
}

func httpListenAndServe(storage usecases.Storage, port int) {
	useCases := usecases.NewUseCases(storage)
	mux := httpintf.NewServeMux(useCases)

	loggerMW := logger.New()
	app := loggerMW.Handler(mux)

	if err := http.ListenAndServe(fmt.Sprintf(`:%d`, port), app); err != nil {
		log.Fatal(err)
	}
}

func createToken(s usecases.Storage, ownerUID string) {
	if ownerUID == `` {
		log.Fatal(`owner uid required to create a token`)
	}

	issuer := security.Issuer{Storage: s}

	t, _, err := issuer.CreateNewToken(context.TODO(), ownerUID, nil, nil)

	if err != nil {
		panic(err)
	}

	fmt.Println(t)
}

func connstr() string {
	connstr, isSet := os.LookupEnv(`DATABASE_URL`)

	if !isSet {
		log.Fatal(`please set "DATABASE_URL" to use the service`)
	}

	return connstr
}
