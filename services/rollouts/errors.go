package rollouts

import "github.com/adamluzsi/frameless"

const (
	ErrMissingFlag        frameless.Error = `feature flag is not provided`
	ErrInvalidURL         frameless.Error = `url value not acceptable`
	ErrInvalidPercentage  frameless.Error = `percentage value not acceptable`
	ErrInvalidFeatureName frameless.Error = `feature name is not acceptable`
)
