package security_test

import (
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestToken(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()

	const (
		UserUID = `The answer is 42`
	)

	token := func(v *testcase.V) *security.Token {
		return &security.Token{
			UserUID:  UserUID,
			IssuedAt: v.I(`IssuedAt`).(time.Time),
			Duration: v.I(`Duration`).(time.Duration),
		}
	}

	s.Describe(`IsValid`, func(s *testcase.Spec) {
		SpecTokenIsValid(token, s)
	})
}

func SpecTokenIsValid(token func(v *testcase.V) *security.Token, s *testcase.Spec) {
	subject := func(v *testcase.V) bool {
		return token(v).IsValid()
	}
	s.When(`the duration is zero`, func(s *testcase.Spec) {
		s.Let(`Duration`, func(v *testcase.V) interface{} {
			return time.Duration(0)
		})

		s.And(`the date when the value is issued is before now`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(v *testcase.V) interface{} {
				return time.Now().Add(-1 * time.Second)
			})

			s.Then(`there will be no expiration and it will be true`, func(t *testing.T, v *testcase.V) {
				require.True(t, subject(v))
			})
		})

		s.And(`the date when the value is issued is in the future`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(v *testcase.V) interface{} {
				return time.Now().Add(42 * time.Second)
			})

			s.Then(`it will be invalid`, func(t *testing.T, v *testcase.V) {
				require.False(t, subject(v))
			})
		})
	})
	s.When(`the duration is limited`, func(s *testcase.Spec) {
		s.Let(`Duration`, func(v *testcase.V) interface{} {
			return time.Minute
		})

		s.And(`the date when the value is issued is before now`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(v *testcase.V) interface{} {
				return time.Now().Add(-1 * time.Second)
			})

			s.And(`not older than the token duration`, func(s *testcase.Spec) {
				s.Let(`IssuedAt`, func(v *testcase.V) interface{} {
					return time.Now().Add(-1*v.I(`Duration`).(time.Duration) + 1*time.Second)
				})

				s.Then(`it is still valid`, func(t *testing.T, v *testcase.V) {
					require.True(t, subject(v))
				})
			})

			s.And(`older than the duration`, func(s *testcase.Spec) {
				s.Let(`IssuedAt`, func(v *testcase.V) interface{} {
					return time.Now().Add(-1*v.I(`Duration`).(time.Duration) + -1*time.Second)
				})

				s.Then(`it is still invalid already`, func(t *testing.T, v *testcase.V) {
					require.False(t, subject(v))
				})
			})

			s.Then(`there will be no expiration and it will be true`, func(t *testing.T, v *testcase.V) {
				require.True(t, subject(v))
			})
		})

		s.And(`the date when the value is issued is in the future`, func(s *testcase.Spec) {
			s.Let(`IssuedAt`, func(v *testcase.V) interface{} {
				return time.Now().Add(42 * time.Second)
			})

			s.Then(`it will be invalid`, func(t *testing.T, v *testcase.V) {
				require.False(t, subject(v))
			})
		})
	})
}
