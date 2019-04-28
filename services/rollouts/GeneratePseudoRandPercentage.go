package rollouts

import (
	"hash/fnv"
	"math/rand"
)

// GeneratePseudoRandPercentageWithFNV1a64 implements pseudo random percentage calculation.
// this is mainly used for pilot enrollments when percentage strategy is used for rollout.
func GeneratePseudoRandPercentageWithFNV1a64(id string, seedSalt int64) (int, error) {
	h := fnv.New64a()

	if _, err := h.Write([]byte(id)); err != nil {
		return 0, err
	}

	seed := int64(h.Sum64()) + seedSalt
	source := rand.NewSource(seed)
	return rand.New(source).Intn(101), nil
}
