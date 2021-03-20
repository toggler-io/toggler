package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/external/resource/caches"
	"github.com/toggler-io/toggler/external/resource/storages"

	"github.com/unrolled/logger"

	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/toggler-io/toggler/domains/toggler"
	"github.com/toggler-io/toggler/external/interface/httpintf"
)

const commandsHelpDescription = `
Commands:
  * http-server, server, s
    - start up http web-server
  * create-token
    - create token for admin user
  * fixtures
    - create fixtures for local development purpose
`

func main() {
	flagSet := flag.NewFlagSet(`toggler`, flag.ContinueOnError)
	dbURL := flagSet.String(`database-url`, ``, `define what url should be used for the db connection. Default value used from ENV[DATABASE_URL].`)
	cacheURL := flagSet.String(`cache-url`, ``, `define what url should be used for the cache connection. default value is taken from ENV[CACHE_URL].`)

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Print(commandsHelpDescription)
			fmt.Println()
		} else {
			log.Print(err.Error())
		}
		os.Exit(0)
	}

	setupDatabaseURL(dbURL)
	setupCacheURL(cacheURL)

	storage, err := storages.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	cache, err := caches.New(*cacheURL, storage)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	switch flagSet.Arg(0) {
	case `create-token`:
		createTokenCMD(flagSet.Args(), storage)

	case `fixtures`:
		fixturesCMD(flagSet.Args(), storage)

	case `http-server`, `server`, `s`:
		httpServerCMD(flagSet.Args(), cache)

	default:
		fmt.Println(`please provide on of the commands`)
		fmt.Printf("\t%s\n", `http-server`)
		fmt.Printf("\t%s\n", `create-token`)
	}

}

func httpServerCMD(args []string, s toggler.Storage) {
	flagSet := flag.NewFlagSet(`http-server`, flag.ExitOnError)
	portConfValue := flagSet.String(`port`, os.Getenv(`PORT`), `set http server port else the env variable "PORT" value will be used.`)

	if err := flagSet.Parse(args[1:]); err != nil {
		log.Println(err)
	}

	httpServer(getPort(*portConfValue), s)
}

func httpServer(port int, storage toggler.Storage) {
	s := makeHTTPServer(storage, port)
	withGracefulShutdown(func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}, func(ctx context.Context) error {
		return s.Shutdown(ctx)
	})
}

func getPort(portConfValue string) int {
	if portConfValue == `` {
		log.Fatal(`http server port is not given (--port/$PORT)`)
	}
	p, err := strconv.Atoi(portConfValue)
	if err != nil {
		log.Fatal(`http server port is not a number`)
	}
	return p
}

func setupDatabaseURL(dbURL *string) {
	if *dbURL == `` {
		return
	}

	if err := os.Setenv(`DATABASE_URL`, *dbURL); err != nil {
		log.Fatal(err)
	}
}

func setupCacheURL(cacheURL *string) {
	if *cacheURL == `` {
		*cacheURL = os.Getenv(`CACHE_URL`)
	}
}

func fixturesCMD(args []string, s toggler.Storage) {
	flagSet := flag.NewFlagSet(`fixtures`, flag.ExitOnError)
	fixtures := flagSet.Bool(`create-fixtures`, false, `create default fixtures for development purpose.`)
	localDevelopmentToken := flagSet.String(`create-unsafe-token`, ``, `create token for local development purpose (don't use in prod)`)

	if err := flagSet.Parse(args[1:]); err != nil {
		log.Println(err)
	}

	if *fixtures {
		createFixtures(s)
	}

	if *localDevelopmentToken != `` {
		createDevelopmentToken(s, *localDevelopmentToken)
	}
}

func createFixtures(s toggler.Storage) {
	uc := toggler.NewUseCases(s)
	ff := release.Flag{Name: `test`}
	ctx := context.Background()
	devEnv := deployment.Environment{Name: "development"}
	_ = uc.Storage.Create(ctx, &devEnv)
	_ = uc.RolloutManager.CreateFeatureFlag(ctx, &ff)
	_ = uc.RolloutManager.SetPilotEnrollmentForFeature(context.Background(), ff.ID, devEnv.ID, `test-public-pilot-id-1`, true)
	_ = uc.RolloutManager.SetPilotEnrollmentForFeature(context.Background(), ff.ID, devEnv.ID, `test-public-pilot-id-2`, false)
}

func makeHTTPServer(storage toggler.Storage, port int) *http.Server {
	useCases := toggler.NewUseCases(storage)
	mux, err := httpintf.NewServeMux(useCases)
	if err != nil {
		log.Fatal(err)
	}

	loggerMW := logger.New()
	app := loggerMW.Handler(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(`:%d`, port),
		Handler: app,
	}

	return server
}

func createTokenCMD(args []string, s toggler.Storage) {
	flagSet := flag.NewFlagSet(`create-token`, flag.ExitOnError)

	flagSet.Usage = func() {
		const format = "Usage of %s: [TOKEN_OWNER_UID]\n"
		_, _ = fmt.Fprintf(flagSet.Output(), format, args[0])
		flagSet.PrintDefaults()
	}

	if err := flagSet.Parse(args[1:]); err != nil {
		log.Fatal(err)
	}

	createToken(s, flagSet.Arg(0))
}

func createToken(s toggler.Storage, ownerUID string) {
	if ownerUID == `` {
		log.Fatal(`owner uid required to create a token`)
	}

	issuer := security.Issuer{Storage: s}

	tStr, _, err := issuer.CreateNewToken(context.Background(), ownerUID, nil, nil)

	if err != nil {
		panic(err)
	}

	fmt.Println(`token:`, tStr)
}

func createDevelopmentToken(s toggler.Storage, tokenSTR string) {
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
