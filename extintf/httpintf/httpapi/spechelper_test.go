package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"

	"github.com/adamluzsi/testcase"
	"github.com/toggler-io/toggler/services/security"
	"github.com/stretchr/testify/require"

	. "github.com/toggler-io/toggler/testing"
)

func UpdateReleaseFlagRolloutPercentage(t *testcase.T, featureFlagName string, rolloutPercentage int) {
	EnsureFlag(t, featureFlagName, rolloutPercentage)
}

func CreateSecurityTokenString(t *testcase.T) string {
	textToken, token, err := security.NewIssuer(GetStorage(t)).CreateNewToken(context.TODO(), GetUniqUserID(t), nil, nil)
	require.Nil(t, err)
	require.NotNil(t, token)
	require.NotEmpty(t, token.SHA512)
	return textToken
}

func IsJsonResponse(t *testcase.T, r *httptest.ResponseRecorder, ptr interface{}) {
	require.Equal(t, "application/json", r.Header().Get(`Content-Type`))
	require.Nil(t, json.NewDecoder(r.Body).Decode(ptr))
}
