package usecases_test

import (
	"github.com/adamluzsi/FeatureFlags/services/rollouts"
	"github.com/adamluzsi/FeatureFlags/services/security"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"testing"

	. "github.com/adamluzsi/FeatureFlags/testing"
)

func TestUseCases_ListFeatureFlags(t *testing.T) {

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`UseCases`, func(t *testcase.T) interface{} {
		return usecases.NewUseCases(t.I(`TestStorage`).(*TestStorage))
	})

	subject := func(t *testcase.T) ([]*rollouts.FeatureFlag, error) {
		uc := t.I(`UseCases`).(*usecases.UseCases)
		return uc.ListFeatureFlags(t.I(`TokenString`).(string))
	}

	onSuccess := func(t *testcase.T) []*rollouts.FeatureFlag {
		ffs, err := subject(t)
		require.Nil(t, err)
		return ffs
	}

	s.When(`token doesn't exist`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			return `The answer is 42`
		})

		s.Before(func(t *testcase.T) {
			require.Nil(t, GetStorage(t).Truncate(security.Token{}))
		})

		s.Then(`we receive back error`, func(t *testcase.T) {
			ffs, err := subject(t)
			require.Error(t, err)
			require.Nil(t, ffs)
		})
	})

	s.When(`token exist`, func(s *testcase.Spec) {
		s.Let(`TokenString`, func(t *testcase.T) interface{} {
			token, err := security.NewIssuer(GetStorage(t)).CreateNewToken(GetUniqUserID(t), nil, nil)
			require.Nil(t, err)
			return token.Token
		})

		s.And(`there are at least one flag in the system`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				rm := rollouts.NewRolloutManager(GetStorage(t))
				require.Nil(t, rm.UpdateFeatureFlagRolloutPercentage(`42`, 42))
			})

			s.Then(`we receive back the feature flag list`, func(t *testcase.T) {
				ffs := onSuccess(t)
				require.Equal(t, 1, len(ffs))
				require.Equal(t, `42`, ffs[0].Name)
			})

		})

		s.And(`there is no flag in the system`, func(s *testcase.Spec) {
			s.Before(func(t *testcase.T) {
				require.Nil(t, GetStorage(t).Truncate(rollouts.FeatureFlag{}))
			})

			s.Then(`we receive back empty list`, func(t *testcase.T) {
				ffs := onSuccess(t)
				require.Equal(t, 0, len(ffs))
			})
		})
	})

}
