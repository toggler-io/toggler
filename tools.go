// +build tools

package toggler

import (
	_ "github.com/go-swagger/go-swagger"
	_ "github.com/golang/mock/gomock"
)

//go:generate mkdir -p .tools
//go:generate go build -o .tools/ github.com/go-swagger/go-swagger/cmd/swagger
//go:generate go build -o .tools/ github.com/golang/mock/mockgen
