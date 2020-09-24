package release

import "github.com/adamluzsi/frameless/consterror"

const (
	ErrNameIsEmpty        consterror.Error = `feature name can't be empty`
	ErrMissingFlag        consterror.Error = `release flag is not provided`
	ErrMissingEnv         consterror.Error = `deployment environment is not provided`
	ErrInvalidAction      consterror.Error = `invalid rollout action`
	ErrFlagAlreadyExist   consterror.Error = `release flag already exist`
	ErrInvalidRequestURL  consterror.Error = `value is not a valid request url`
	ErrInvalidPercentage  consterror.Error = `percentage value not acceptable`
	ErrMissingRolloutPlan consterror.Error = `release rollout plan is not provided`
)
