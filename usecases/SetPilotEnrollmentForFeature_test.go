package usecases_test

import (
	"math/rand"
	"testing"

	. "github.com/adamluzsi/FeatureFlags/testing"
	"github.com/adamluzsi/testcase"
	"github.com/stretchr/testify/require"
)

func TestUseCases_SetPilotEnrollmentForFeature(t *testing.T) {
	s := testcase.NewSpec(t)
	SetupSpecCommonVariables(s)
	SetupSpec(s)
	s.Parallel()

	subject := func(t *testcase.T) error {
		return GetProtectedUsecases(t).SetPilotEnrollmentForFeature(
			GetFeatureFlagName(t),
			GetExternalPilotID(t),
			t.I(`expected enrollment`).(bool),
		)
	}

	s.Let(`expected enrollment`, func(t *testcase.T) interface{} {
		return rand.Intn(2) == 0
	})

	s.Then(`it will set percentage`, func(t *testcase.T) {
		require.Nil(t, subject(t))

		flag, err := GetStorage(t).FindFlagByName(GetFeatureFlagName(t))
		require.Nil(t, err)

		pilot, err := GetStorage(t).FindFlagPilotByExternalPilotID(flag.ID, GetExternalPilotID(t))
		require.Nil(t, err)
		require.NotNil(t, pilot)
		require.Equal(t, t.I(`expected enrollment`).(bool), pilot.Enrolled)
	})

}
