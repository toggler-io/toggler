package caches

import "github.com/adamluzsi/frameless/reflects"

func getT(ent interface{}) interface{} {
	return reflects.BaseValueOf(ent).Interface()
}
