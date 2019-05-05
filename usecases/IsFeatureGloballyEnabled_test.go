package usecases_test

import (
	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/FeatureFlags/usecases"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestUseCases_IsFeatureGloballyEnabled(t *testing.T) {

	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	s.Parallel()

	s.Let(`UseCases`, func(v *testcase.V) interface{} {
		return usecases.NewUseCases(v.I(`Storage`).(*Storage))
	})

	subject := func(v *testcase.V) (bool, error) {
		uc := v.I(`UseCases`).(*usecases.UseCases)
		return uc.IsFeatureGloballyEnabled(v.I(`FeatureName`).(string))
	}

	isEnrolled := func(t *testing.T, v *testcase.V) bool {
		enrolled, err := subject(v)
		require.Nil(t, err)
		return enrolled
	}

	s.When(`flag is already configured`, func(s *testcase.Spec) {
		s.Before(func(t *testing.T, v *testcase.V) {
			require.Nil(t, v.I(`UseCases`).(*usecases.UseCases).
				UpdateFeatureFlagRolloutPercentage(v.I(`FeatureName`).(string),
					v.I(`percentage`).(int)))
		})

		s.And(`with global rollout (100%)`, func(s *testcase.Spec) {
			s.Let(`percentage`, func(v *testcase.V) interface{} { return int(100) })

			s.Then(`the feature will be reportad to be globally enabled`, func(t *testing.T, v *testcase.V) {
				require.True(t, isEnrolled(t, v))
			})
		})

		s.And(`with less than 100%`, func(s *testcase.Spec) {
			s.Let(`percentage`, func(v *testcase.V) interface{} { return rand.Intn(100) })

			s.Then(`it will report that feature is currently not available globally`, func(t *testing.T, v *testcase.V) {
				require.False(t, isEnrolled(t, v))
			})
		})
	})

}
