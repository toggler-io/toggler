package release

import (
	"hash/fnv"
	"math/rand"
)

// PseudoRandPercentageGenerator implements pseudo random percentage calculations with different algorithms.
// This is mainly used for pilot enrollments when percentage strategy is used for rollout.
type PseudoRandPercentageGenerator struct{}

// FNV1a64 implements pseudo random percentage calculation with FNV-1a64.
func (g PseudoRandPercentageGenerator) FNV1a64(id string, seedSalt int64) (int, error) {
	h := fnv.New64a()

	if _, err := h.Write([]byte(id)); err != nil {
		return 0, err
	}

	seed := int64(h.Sum64()) + seedSalt
	source := rand.NewSource(seed)
	return rand.New(source).Intn(101), nil
}
