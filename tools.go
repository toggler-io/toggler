// +build tools

package toggler

//go:generate go install github.com/golang/mock/mockgen
//go:generate go install github.com/go-swagger/go-swagger/cmd/swagger

import (
	_ "github.com/go-swagger/go-swagger"
	_ "github.com/golang/mock/gomock"
	_ "github.com/mjibson/esc"
)
