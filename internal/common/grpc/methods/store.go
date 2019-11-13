package methods

import (
	"fmt"
	"p2pderivatives-server/internal/common/grpc/pbbase"

	"github.com/golang/protobuf/proto"
	validator "github.com/mwitkow/go-proto-validators"
	"google.golang.org/grpc"

	"github.com/jhump/protoreflect/grpcreflect"
)

var methodStore map[string]*pbbase.OptionBase

func init() {
	//registers the validator.proto file to use in protobuf definition files.
	fieldValidator := &validator.FieldValidator{}
	bytes, _ := fieldValidator.Descriptor()
	proto.RegisterFile("validator.proto", bytes)
}

// Init initializes the module by recording the methods on the given server
// instance that use an option.
func Init(svr *grpc.Server) {
	sds, _ := grpcreflect.LoadServiceDescriptors(svr)
	methodStore = make(map[string]*pbbase.OptionBase)
	for _, sd := range sds {
		for _, md := range sd.GetMethods() {
			opts := md.GetMethodOptions()

			val, err := proto.GetExtension(opts, pbbase.E_OptionBase)
			if err == nil {
				option, ok := val.(*pbbase.OptionBase)
				if ok {
					methodStore[fmt.Sprintf("/%s/%s", sd.GetFullyQualifiedName(), md.GetName())] = option
				}
			}
		}
	}
}

// IsIgnoreTokenVerify returns whether the given method name has the
// "IgnoreTokenVerify" option.
func IsIgnoreTokenVerify(methodName string) bool {
	val, ok := methodStore[methodName]
	if ok {
		return val.IgnoreTokenVerify
	}
	return false
}

// TxOption returns the value of the "TxOption" option for the requested
// method. If the method doesn't have the option, R/W option is returned.
func TxOption(methodName string) pbbase.TxOption {
	val, ok := methodStore[methodName]
	if ok {
		return val.TxOption
	}
	return pbbase.TxOption_ReadWrite
}
