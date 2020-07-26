// +build tools

// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	_ "github.com/golang/mock/gomock"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/mwitkow/go-proto-validators/protoc-gen-govalidators"
	_ "gotest.tools/gotestsum"
)
