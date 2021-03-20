package caches

import (
	"fmt"

	"github.com/toggler-io/toggler/domains/deployment"
	"github.com/toggler-io/toggler/domains/release"
	"github.com/toggler-io/toggler/domains/security"
)


type EntityTypeNameMappingFunc func(T interface{}) string

func defaultEntityTypeNameMapping(T interface{}) string {
	switch T.(type) {
	case release.Flag:
		return `release_flag`
	case release.ManualPilot:
		return `release_pilot`
	case release.Rollout:
		return `release_rollout`
	case deployment.Environment:
		return `deployment_environment`
	case CachedQuery:
		return `cached_query`
	case security.Token:
		return `security_token`
	default:
		panic(fmt.Sprintf(`unknown type: %T`, T))
	}
}
