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
	"os/signal"
	"strconv"
	"syscall"
	"time"

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
	localDevelopmentToken := flagSet.String(`create-development-token`, ``, `create token for local development purpose only!`)
	dbURL := flagSet.String(`database-url`, ``, `define what url should be used for the db connection. Default value used from ENV[DATABASE_URL].`)
	cacheURL := flagSet.String(`cache-url`, ``, `define what url should be used for the cache connection. default value is taken from ENV[CACHE_URL].`)
	cacheTTL := flagSet.Duration(`cache-ttl`, 30*time.Minute, `define the time-to-live duration for the cached objects (if cache used)`)

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	setupDatabaseURL(dbURL)
	setupCacheURL(cacheURL, cacheTTL)

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

	if err := cache.SetTimeToLiveForValuesToCache(*cacheTTL); err != nil {
		log.Fatal(err)
	}

	if *fixtures {
		createFixtures(storage)
	}

	if *localDevelopmentToken != `` {
		createDevelopmentToken(storage, *localDevelopmentToken)
	}

	switch *cmd {
	case `http-server`:
		port, err := strconv.Atoi(*portConfValue)
		if err != nil {
			panic(err)
		}

		s := makeHTTPServer(cache, port)
		withGracefulShutdown(func() {
			if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}, func(ctx context.Context) error {
			return s.Shutdown(ctx)
		})

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

func setupCacheURL(cacheURL *string, cacheTTL *time.Duration) {
	if *cacheURL == `` {
		*cacheURL = os.Getenv(`CACHE_URL`)
	}

	serializedTTL, isSet := os.LookupEnv(`CACHE_TTL`)

	if isSet {
		d, err := time.ParseDuration(serializedTTL)
		if err != nil {
			log.Println(`parsing CACHE_TTL failed`)
			log.Fatal(err)
		}
		*cacheTTL = d
	}
}

func createFixtures(s usecases.Storage) {
	useCases := usecases.NewUseCases(s)
	issuer := security.Issuer{Storage: s}

	tstr, t, err := issuer.CreateNewToken(context.Background(), `testing`, nil, nil)
	if err != nil {
		panic(err)
	}
	defer s.DeleteByID(context.Background(), *t, t.ID)

	pu, err := useCases.ProtectedUsecases(context.Background(), tstr)

	if err != nil {
		panic(err)
	}

	ff := rollouts.FeatureFlag{Name: `test`}
	_ = pu.CreateFeatureFlag(context.TODO(), &ff)
	_ = pu.SetPilotEnrollmentForFeature(context.Background(), ff.ID, `test-public-pilot-id-1`, true)
	_ = pu.SetPilotEnrollmentForFeature(context.Background(), ff.ID, `test-public-pilot-id-2`, false)
}

func makeHTTPServer(storage usecases.Storage, port int) *http.Server {
	useCases := usecases.NewUseCases(storage)
	mux := httpintf.NewServeMux(useCases)

	loggerMW := logger.New()
	app := loggerMW.Handler(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(`:%d`, port),
		Handler: app,
	}

	return server
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

func createDevelopmentToken(s usecases.Storage, tokenSTR string) {
	defer func() {
		fmt.Println(`WARNING - you created a non random token for local development purpose`)
		fmt.Println(`if this is a production environment, it is advised to delete this token immediately`)
	}()
	
	ctx := context.Background()

	issuer := security.Issuer{Storage: s}
	dk := security.NewDoorkeeper(s)

	valid, err := dk.VerifyTextToken(ctx, tokenSTR)
	if err != nil {
		panic(err)
	}

	if valid {
		return
	}

	_, t, err := issuer.CreateNewToken(ctx, `developer`, nil, nil)

	if err != nil {
		panic(err)
	}

	sha512hex, err := security.ToSHA512Hex(tokenSTR)

	if err != nil {
		panic(err)
	}

	t.SHA512 = sha512hex

	if err := s.Update(ctx, t); err != nil {
		panic(err)
	}
}

func withGracefulShutdown(wrk func(), shutdown func(context.Context) error) {

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go wrk()
	<-done

	log.Println(`toggler graceful shutdown - BEGIN`)

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	if err := shutdown(ctx); err != nil {
		log.Fatalln(err)
	}

	log.Println(`toggler graceful shutdown - FINISH`)

}
