package contracts

import (
	"sync"
	"testing"
	"time"

	"github.com/adamluzsi/frameless/contracts"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/fixtures"
	"github.com/stretchr/testify/require"
	"github.com/toggler-io/toggler/domains/security"
)

type TokenStorage struct {
	Subject func(testing.TB) security.TokenStorage
	contracts.FixtureFactory
}

func (spec TokenStorage) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec TokenStorage) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec TokenStorage) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`TokenStorage`, func(s *testcase.Spec) {
		T := security.Token{}
		testcase.RunContract(s,
			contracts.Creator{
				T:              T,
				Subject:        func(tb testing.TB) contracts.CRD { return spec.Subject(tb) },
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Finder{
				T:              T,
				Subject:        func(tb testing.TB) contracts.CRD { return spec.Subject(tb) },
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Updater{
				T:              T,
				Subject:        func(tb testing.TB) contracts.UpdaterSubject { return spec.Subject(tb) },
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Deleter{
				T:              T,
				Subject:        func(tb testing.TB) contracts.CRD { return spec.Subject(tb) },
				FixtureFactory: spec.FixtureFactory,
			},
			TokenFinder{
				Subject:        func(tb testing.TB) security.TokenStorage { return spec.Subject(tb) },
				FixtureFactory: spec.FixtureFactory,
			},
			contracts.Publisher{T: T,
				Subject: func(tb testing.TB) contracts.PublisherSubject {
					return spec.Subject(tb)
				},
				FixtureFactory: spec.FixtureFactory,
			},
		)
	})
}

type TokenFinder struct {
	Subject func(testing.TB) security.TokenStorage
	contracts.FixtureFactory
}

func (spec TokenFinder) Test(t *testing.T) {
	spec.Spec(t)
}

func (spec TokenFinder) Benchmark(b *testing.B) {
	spec.Spec(b)
}

func (spec TokenFinder) Spec(tb testing.TB) {
	testcase.NewSpec(tb).Describe(`TokenFinder`, func(s *testcase.Spec) {
		uid := s.Let(`uid`, func(t *testcase.T) interface{} { return fixtures.Random.String() })
		tokenSHA := s.Let(`token SHA512`, func(t *testcase.T) interface{} {
			return `the answer is 42`
		})
		token := s.Let(`token object`, func(t *testcase.T) interface{} {
			return &security.Token{
				OwnerUID: uid.Get(t).(string),
				SHA512:   tokenSHA.Get(t).(string),
				IssuedAt: fixtures.Random.Time().UTC(),
				Duration: 1 * time.Second,
			}
		})
		tokenGet := func(t *testcase.T) *security.Token {
			return token.Get(t).(*security.Token)
		}

		storage := s.Let(`storage`, func(t *testcase.T) interface{} {
			return spec.Subject(t)
		})
		storageGet := func(t *testcase.T) security.TokenStorage {
			return storage.Get(t).(security.TokenStorage)
		}

		deleteAllTokenOnce := &sync.Once{}
		s.Before(func(t *testcase.T) {
			deleteAllTokenOnce.Do(func() {
				contracts.DeleteAllEntity(t, storageGet(t), spec.Context())
			})
		})

		s.Describe(`.FindTokenBySHA512Hex`, func(s *testcase.Spec) {
			subject := func(t *testcase.T) (*security.Token, error) {
				return storageGet(t).FindTokenBySHA512Hex(spec.Context(), tokenSHA.Get(t).(string))
			}

			s.When(`no token stored in the storage yet`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					count, err := iterators.Count(storageGet(t).FindAll(spec.Context()))
					require.Nil(t, err)
					require.Equal(t, 0, count)
				})

				s.Then(`it will return nil token without any error`, func(t *testcase.T) {
					token, err := subject(t)
					require.Nil(t, err)
					require.Nil(t, token)
				})
			})

			s.When(`token is stored in the storage already`, func(s *testcase.Spec) {
				s.Before(func(t *testcase.T) {
					contracts.CreateEntity(t, storageGet(t), spec.Context(), tokenGet(t))
				})

				s.Then(`token will be retrieved`, func(t *testcase.T) {
					actual, err := subject(t)
					expected := tokenGet(t)

					require.Nil(t, err)
					require.NotNil(t, actual)
					require.Equal(t, expected, actual)
				})
			})

		})

	})
}
