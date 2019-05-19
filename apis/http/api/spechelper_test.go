package api_test

import (
	"encoding/json"
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"net/http/httptest"

	. "github.com/adamluzsi/FeatureFlags/testing"
)


func UpdateFeatureFlagRolloutPercentage(t *testcase.T, featureFlagName string, rolloutPercentage int) {
	rm := rollouts.NewRolloutManager(GetStorage(t))
	require.Nil(t, rm.UpdateFeatureFlagRolloutPercentage(
		featureFlagName,
		rolloutPercentage,
	))
}

func CreateSecurityTokenString(t *testcase.T) string {
	token, err := security.NewIssuer(GetStorage(t)).CreateNewToken(GetUniqUserID(t), nil, nil)
	require.Nil(t, err)
	return token.Token
}

func IsJsonRespone(t *testcase.T, r *httptest.ResponseRecorder, ptr interface{}) {
	require.Equal(t, "application/json", r.Header().Get(`Content-Type`))
	require.Nil(t, json.NewDecoder(r.Body).Decode(ptr))
}