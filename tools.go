// +build tools

package toggler

//go:generate go install github.com/golang/mock/mockgen

import (
	_ "github.com/golang/mock/gomock"
	_ "github.com/mjibson/esc"
)
