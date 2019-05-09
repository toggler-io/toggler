package security

import (
	"github.com/adamluzsi/frameless/resources/specs"
)

type Storage interface {
	specs.Save
	specs.FindByID
	specs.Truncate
	specs.DeleteByID
	specs.Update
}

type TokenFinder interface {
	FindByTokenHashSum(hashsum string) (*Token, error)
}
