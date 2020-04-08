package testing

import (
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

	"github.com/toggler-io/toggler/domains/security"
)

const (
	TokenLetVar     = `testing token`
	TokenTextLetVar = `testing token as text`
	TokenTextPtrLetVar = `testing token text pointer`
)

func init() {
	setups = append(setups, func(s *testcase.Spec) {
		s.Let(TokenLetVar, func(t *testcase.T) interface{} {
			textToken, objectToken := CreateToken(t, GetUniqueUserID(t))
			*(t.I(TokenTextPtrLetVar).(*string)) = textToken
			return objectToken
		})

		s.Let(TokenTextPtrLetVar, func(t *testcase.T) interface{} {
			var textToken string
			return &textToken
		})

		s.Let(TokenTextLetVar, func(t *testcase.T) interface{} {
			t.I(TokenLetVar) // trigger *TextToken Setup
			return *(t.I(TokenTextPtrLetVar).(*string))
		})
	})
}

func GetTextToken(t *testcase.T) string {
	return t.I(TokenTextLetVar).(string)
}

func GetToken(t *testcase.T) *security.Token {
	return t.I(TokenLetVar).(*security.Token)
}

func CreateToken(t *testcase.T, tokenOwner string) (string, *security.Token) {
	issuer := security.NewIssuer(ExampleStorage(t))
	textToken, token, err := issuer.CreateNewToken(GetContext(t), tokenOwner, nil, nil)
	require.Nil(t, err)
	return textToken, token
}
