package security_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/stretchr/testify/require"

	. "github.com/adamluzsi/toggler/testing"
)

func TestIssuer(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`issuer`, func(t *testcase.T) interface{} {
		return security.NewIssuer(GetStorage(t))
	})

	SpecIssuerCreateNewToken(s)
	SpecIssuerRevokeToken(s)

}

func SpecIssuerRevokeToken(s *testcase.Spec) {
	s.Describe(`RevokeToken`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) error {
			token, _ := t.I(`Token`).(*security.Token)
			issuer := t.I(`issuer`).(*security.Issuer)
			return issuer.RevokeToken(context.TODO(), token)
		}

		s.When(`token exists`, func(s *testcase.Spec) {
			s.Let(`Token`, func(t *testcase.T) interface{} {
				issuer := t.I(`issuer`).(*security.Issuer)
				token, err := issuer.CreateNewToken(context.TODO(), GetUniqUserID(t), nil, nil)
				require.Nil(t, err)
				return token
			})

			s.Before(func(t *testcase.T) {
				require.NotNil(t, t.I(`Token`))
			})

			s.Then(`token will be revoked`, func(t *testcase.T) {
				require.Nil(t, subject(t))
				token := t.I(`Token`).(*security.Token)

				dk := security.NewDoorkeeper(GetStorage(t))
				valid, err := dk.VerifyTokenString(token.Token)
				require.Nil(t, err)
				require.False(t, valid)
			})
		})
	})
}

func SpecIssuerCreateNewToken(s *testcase.Spec) {
	s.Describe(`CreateNewToken`, func(s *testcase.Spec) {
		subject := func(t *testcase.T) (*security.Token, error) {
			issuer := t.I(`issuer`).(*security.Issuer)
			userUID := t.I(`userUID`).(string)
			issueAt, _ := t.I(`issueAt`).(*time.Time)
			duration, _ := t.I(`duration`).(*time.Duration)
			return issuer.CreateNewToken(context.TODO(), userUID, issueAt, duration)
		}
		onSuccess := func(t *testcase.T) *security.Token {
			token, err := subject(t)
			require.Nil(t, err)
			require.NotNil(t, token)
			return token
		}
		onFailure := func(t *testcase.T) error {
			token, err := subject(t)
			require.Nil(t, token)
			require.NotNil(t, err)
			return err
		}
		givenWeHaveValidParameters := func(s *testcase.Spec) {
			s.Let(`userUID`, func(t *testcase.T) interface{} { return ExampleUniqUserID() })
			s.Let(`issueAt`, func(t *testcase.T) interface{} { ia := time.Now().UTC(); return &ia })
			s.Let(`duration`, func(t *testcase.T) interface{} { d := 42 * time.Hour; return &d })
		}
		s.When(`all parameter acceptable`, func(s *testcase.Spec) {
			givenWeHaveValidParameters(s)

			s.Then(`we receive a token back`, func(t *testcase.T) {
				token := onSuccess(t)
				require.Equal(t, t.I(`userUID`).(string), token.OwnerUID)
				require.Equal(t, t.I(`issueAt`).(*time.Time), &token.IssuedAt)
				require.Equal(t, t.I(`duration`).(*time.Duration), &token.Duration)
			})

			s.Then(`the token generated with a long token key`, func(t *testcase.T) {
				token := onSuccess(t)
				require.True(t, 128 <= len(token.Token))
			})

			s.Then(`token is stored in the storage`, func(t *testcase.T) {
				t1 := onSuccess(t)
				t2 := security.Token{}

				found, err := t.I(`TestStorage`).(*TestStorage).FindByID(context.Background(), &t2, t1.ID)
				require.Nil(t, err)
				require.True(t, found)
				require.Equal(t, t1, &t2)
			})

			s.Then(`each time a token is created, it will be uniq`, func(t *testcase.T) {
				issuer := t.I(`issuer`).(*security.Issuer)
				issueAt := t.I(`issueAt`).(*time.Time)
				duration := t.I(`duration`).(*time.Duration)

				var last string
				for i := 0; i < 1024; i++ {
					token, err := issuer.CreateNewToken(context.TODO(), strconv.Itoa(i), issueAt, duration)
					require.Nil(t, err)
					require.NotNil(t, token)

					if last == "" {
						last = token.Token
						continue
					}

					t.Log(token.Token)
					require.NotEqual(t, last, token.Token)
				}
			})
		})
		s.When(`userUID is empty`, func(s *testcase.Spec) {
			givenWeHaveValidParameters(s)
			s.Let(`userUID`, func(t *testcase.T) interface{} { return `` })

			s.Then(`we receive error because empty userUID is not acceptable`, func(t *testcase.T) {
				require.Error(t, onFailure(t))
			})
		})
		s.When(`duration is not provided`, func(s *testcase.Spec) {
			givenWeHaveValidParameters(s)
			s.Let(`duration`, func(t *testcase.T) interface{} { return nil })

			s.Then(`it will create a token that can't expire`, func(t *testcase.T) {
				token := onSuccess(t)
				require.False(t, token.IsExpirable())
			})
		})
	})
}
