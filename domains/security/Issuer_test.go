package security_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/adamluzsi/frameless/fixtures"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/security"

	. "github.com/toggler-io/toggler/testing"
)

func TestIssuer(t *testing.T) {
	s := testcase.NewSpec(t)
	SetUp(s)
	s.Parallel()

	s.Let(`issuer`, func(t *testcase.T) interface{} {
		return security.NewIssuer(ExampleStorage(t))
	})

	s.Describe(`CreateNewToken`, SpecIssuerCreateNewToken)
	s.Describe(`RevokeToken`, SpecIssuerRevokeToken)
}

func SpecIssuerRevokeToken(s *testcase.Spec) {
	var subject = func(t *testcase.T) error {
		token, _ := t.I(`token`).(*security.Token)
		issuer := t.I(`issuer`).(*security.Issuer)
		return issuer.RevokeToken(GetContext(t), token)
	}

	s.When(`token exists`, func(s *testcase.Spec) {
		s.Let(`token`, func(t *testcase.T) interface{} {
			issuer := t.I(`issuer`).(*security.Issuer)
			_, token, err := issuer.CreateNewToken(GetContext(t), ExampleUniqueUserID(t), nil, nil)
			require.Nil(t, err)

			t.Defer(ExampleStorage(t).DeleteByID, GetContext(t), security.Token{}, token.ID)
			return token
		})

		s.Before(func(t *testcase.T) {
			require.NotNil(t, t.I(`token`))
		})

		s.Then(`token will be revoked`, func(t *testcase.T) {
			require.Nil(t, subject(t))
			token := t.I(`token`).(*security.Token)

			dk := security.NewDoorkeeper(ExampleStorage(t))
			valid, err := dk.VerifyTextToken(GetContext(t), token.SHA512)
			require.Nil(t, err)
			require.False(t, valid)
		})
	})
}

func SpecIssuerCreateNewToken(s *testcase.Spec) {
	var subject = func(t *testcase.T) (string, *security.Token, error) {
		issuer := t.I(`issuer`).(*security.Issuer)
		userUID := t.I(`userUID`).(string)
		issueAt, _ := t.I(`issueAt`).(*time.Time)
		duration, _ := t.I(`duration`).(*time.Duration)
		return issuer.CreateNewToken(GetContext(t), userUID, issueAt, duration)
	}
	onSuccess := func(t *testcase.T) *security.Token {
		textToken, token, err := subject(t)
		require.Nil(t, err)
		require.NotNil(t, token)
		t.Defer(ExampleStorage(t).DeleteByID, GetContext(t), *token, token.ID)
		hashed, err := security.ToSHA512Hex(textToken)
		require.Nil(t, err)
		require.Equal(t, hashed, token.SHA512)
		return token
	}
	onFailure := func(t *testcase.T) error {
		textToken, token, err := subject(t)
		require.Empty(t, textToken)
		require.Nil(t, token)
		require.NotNil(t, err)
		return err
	}
	givenWeHaveValidParameters := func(s *testcase.Spec) {
		s.Let(`userUID`, func(t *testcase.T) interface{} {
			return fixtures.Random.String()
		})
		s.Let(`issueAt`, func(t *testcase.T) interface{} {
			ia := fixtures.Random.Time()
			return &ia
		})
		s.Let(`duration`, func(t *testcase.T) interface{} {
			d := time.Duration(fixtures.Random.IntN(42)) * time.Hour
			return &d
		})
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
			require.True(t, 128 <= len(token.SHA512))
		})

		s.Then(`token is stored in the storage`, func(t *testcase.T) {
			t1 := onSuccess(t)
			t2 := security.Token{}

			found, err := ExampleStorage(t).FindByID(GetContext(t), &t2, t1.ID)
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
				_, token, err := issuer.CreateNewToken(GetContext(t), strconv.Itoa(i), issueAt, duration)
				require.Nil(t, err)
				require.NotNil(t, token)

				if last == "" {
					last = token.SHA512
					continue
				}

				require.NotEqual(t, last, token.SHA512)
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
}
