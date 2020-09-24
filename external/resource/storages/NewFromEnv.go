package storages

import (
	"net/url"
	"os"

	"github.com/adamluzsi/frameless/consterror"

	"github.com/toggler-io/toggler/domains/toggler"
)

const ErrRDSEngineNotSet consterror.Error = `RDS_ENGINE environment variable not set.
You need to set the environment variable by hand to the engine being used.
e.g.: RDS_ENGINE=postgres`

const ErrNewFromErrNotPossible consterror.Error = `storage initialization from environment variables failed.
Missing variables from your environment for storage initialization.
Please set "DATABASE_URL" in your environment variable to solve this.
e.g.: DATABASE_URL=memory`

func NewFromEnv() (toggler.Storage, error) {
	if databaseURL, ok := os.LookupEnv(`DATABASE_URL`); ok {
		return New(databaseURL)
	}

	if _, ok := os.LookupEnv(`RDS_DB_NAME`); ok {
		if _, ok := os.LookupEnv(`RDS_ENGINE`); !ok {
			return nil, ErrRDSEngineNotSet
		}

		vs, err := url.ParseQuery(os.Getenv(`RDS_ENGINE_OPTS`))
		if err != nil {
			return nil, err
		}

		return New((&url.URL{
			Scheme:   os.Getenv(`RDS_ENGINE`),
			Path:     os.Getenv(`RDS_DB_NAME`),
			Host:     os.Getenv(`RDS_HOSTNAME`) + `:` + os.Getenv(`RDS_PORT`),
			RawQuery: vs.Encode(),
			User: url.UserPassword(
				os.Getenv(`RDS_USERNAME`),
				os.Getenv(`RDS_PASSWORD`),
			),
		}).String())
	}

	return nil, ErrNewFromErrNotPossible
}
