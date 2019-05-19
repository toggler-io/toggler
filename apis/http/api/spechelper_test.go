package api_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"

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