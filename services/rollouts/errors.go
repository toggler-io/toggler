package rollouts

import "github.com/adamluzsi/frameless"

const (
	ErrNameIsEmpty       frameless.Error = `feature name can't be empty`
	ErrMissingFlag       frameless.Error = `feature flag is not provided`
	ErrInvalidAction     frameless.Error = `invalid rollout action`
	ErrFlagAlreadyExist  frameless.Error = `feature flag already exist`
	ErrInvalidRequestURL frameless.Error = `value is not a valid request url`
	ErrInvalidPercentage frameless.Error = `percentage value not acceptable`
)
