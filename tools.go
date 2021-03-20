// +build tools

//go:generate mkdir -p .tools
package toggler

//go:generate go build -o .tools/ github.com/go-swagger/go-swagger/cmd/swagger
import (
	_ "github.com/go-swagger/go-swagger"

	//go:generate go build -o .tools/ github.com/golang/mock/mockgen
	_ "github.com/golang/mock/gomock"

	//go:generate go build -o .tools/ github.com/mjibson/esc
	_ "github.com/mjibson/esc"
)
