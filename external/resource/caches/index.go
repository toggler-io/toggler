package caches

import "fmt"

type index map[string]struct{}

func (s index) Add(v interface{}) {
	s[s.toKey(v)] = struct{}{}
}

func (s index) Has(v interface{}) bool {
	_, ok := s[s.toKey(v)]
	return ok
}

func (s index) toKey(v interface{}) string {
	return fmt.Sprintf(`%#v`, v)
}
