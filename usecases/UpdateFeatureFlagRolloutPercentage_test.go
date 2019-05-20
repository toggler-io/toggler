package usecases_test

import (
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestUseCases_UpdateFeatureFlagRolloutPercentage(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	SetupSpec(s)
	s.Parallel()

	subject := func(t *testcase.T) error {
		return GetUseCases(t).UpdateFeatureFlagRolloutPercentage(
			t.I(`TokenString`).(string),
			GetFeatureFlagName(t),
			GetRolloutPercentage(t),
		)
	}

	s.Let(`RolloutPercentage`, func(t *testcase.T) interface{} { return rand.Intn(128) + 1 })

	s.When(`caller have invalid token`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} { return `invalid token` })

		s.Then(`it will be rejected`, func(t *testcase.T) {
			require.Equal(t, usecases.ErrInvalidToken, subject(t))
		})
	})

	s.When(`caller have a valid token`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			return CreateToken(t, `rollout manager uniq id`).Token
		})

		s.Then(`it will set percentage`, func(t *testcase.T) {
			require.Nil(t, subject(t))
			flag, err := GetStorage(t).FindFlagByName(GetFeatureFlagName(t))
			require.Nil(t, err)
			require.NotNil(t, flag)
			require.Equal(t, GetRolloutPercentage(t), flag.Rollout.Strategy.Percentage)
		})
	})

}
