package usecases_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/adamluzsi/testcase"
	. "github.com/toggler-io/toggler/testing"
	"github.com/stretchr/testify/require"
)

func TestUseCases_SetPilotEnrollmentForFeature(t *testing.T) {
	s := testcase.NewSpec(t)
	s.Parallel()
	SetUp(s)

	s.Before(func(t *testcase.T) {
		require.Nil(t, GetProtectedUsecases(t).CreateFeatureFlag(context.Background(), ExampleReleaseFlag(t)))
	})

	subject := func(t *testcase.T) error {
		return GetProtectedUsecases(t).SetPilotEnrollmentForFeature(context.Background(), ExampleReleaseFlag(t).ID, GetExternalPilotID(t), t.I(`expected enrollment`).(bool), )
	}

	s.Let(`expected enrollment`, func(t *testcase.T) interface{} {
		return rand.Intn(2) == 0
	})

	s.Then(`it will set enrollment`, func(t *testcase.T) {
		require.Nil(t, subject(t))

		pilot, err := ExampleStorage(t).FindReleaseFlagPilotByPilotExternalID(context.Background(), ExampleReleaseFlag(t).ID, GetExternalPilotID(t))
		require.Nil(t, err)
		require.NotNil(t, pilot)
		require.Equal(t, t.I(`expected enrollment`).(bool), pilot.Enrolled)
	})

}
