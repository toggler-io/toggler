package security_test

import (
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/domains/security"
	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	const (
		UserUID = `The answer is 42`
	)

	token := func(t *testcase.T) *security.Token {
		return &security.Token{
			OwnerUID: UserUID,
			IssuedAt: t.I(`IssuedAt`).(time.Time),
			Duration: t.I(`Duration`).(time.Duration),
		}
	}

	s.Describe(`IsValid`, func(s *testcase.Spec) {
		SpecTokenIsValid(token, s)
	})

	s.Describe(`IsExpirable`, func(s *testcase.Spec) {
		SpecTokenIsExpirable(token, s)
	})

}

func SpecTokenIsValid(token func(t *testcase.T) *security.Token, s *testcase.Spec) {
	subject := func(t *testcase.T) bool {
		return token(t).IsValid()
	}

	s.When(`the duration is zero`, func(s *testcase.Spec) {
		s.Let(`Duration`, func(t *testcase.T) interface{} {
			return time.Duration(0)
		})

		s.And(`the date when the value is issued is before now`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(t *testcase.T) interface{} {
				return time.Now().Add(-1 * time.Second)
			})

			s.Then(`there will be no expiration and it will be true`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})

		s.And(`the date when the value is issued is in the future`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(t *testcase.T) interface{} {
				return time.Now().Add(42 * time.Second)
			})

			s.Then(`it will be invalid`, func(t *testcase.T) {
				require.False(t, subject(t))
			})
		})
	})
	s.When(`the duration is limited`, func(s *testcase.Spec) {
		s.Let(`Duration`, func(t *testcase.T) interface{} {
			return time.Minute
		})

		s.And(`the date when the value is issued is before now`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(t *testcase.T) interface{} {
				return time.Now().Add(-1 * time.Second)
			})

			s.And(`not older than the token duration`, func(s *testcase.Spec) {
				s.Let(`IssuedAt`, func(t *testcase.T) interface{} {
					return time.Now().Add(-1*t.I(`Duration`).(time.Duration) + 1*time.Second)
				})

				s.Then(`it is still valid`, func(t *testcase.T) {
					require.True(t, subject(t))
				})
			})

			s.And(`older than the duration`, func(s *testcase.Spec) {
				s.Let(`IssuedAt`, func(t *testcase.T) interface{} {
					return time.Now().Add(-1*t.I(`Duration`).(time.Duration) + -1*time.Second)
				})

				s.Then(`it is still invalid already`, func(t *testcase.T) {
					require.False(t, subject(t))
				})
			})

			s.Then(`there will be no expiration and it will be true`, func(t *testcase.T) {
				require.True(t, subject(t))
			})
		})

		s.And(`the date when the value is issued is in the future`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(t *testcase.T) interface{} {
				return time.Now().Add(42 * time.Second)
			})

			s.Then(`it will be invalid`, func(t *testcase.T) {
				require.False(t, subject(t))
			})
		})
	})
}

func SpecTokenIsExpirable(token func(t *testcase.T) *security.Token, s *testcase.Spec) {
	subject := func(t *testcase.T) bool {
		return token(t).IsExpirable()
	}

	s.Let(`IssuedAt`, func(t *testcase.T) interface{} { return time.Now().UTC() })

	s.When(`the duration is zero`, func(s *testcase.Spec) {
		s.Let(`Duration`, func(t *testcase.T) interface{} { return time.Duration(0) })

		s.Then(`there will be good forever, so will not expire`, func(t *testcase.T) {
			require.False(t, subject(t))
		})
	})

	s.When(`the duration is limited`, func(s *testcase.Spec) {
		s.Let(`Duration`, func(t *testcase.T) interface{} { return time.Minute })

		s.Then(`it will be expirable`, func(t *testcase.T) {
			require.True(t, subject(t))
		})
	})
}
