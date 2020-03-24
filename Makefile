setup: install gen deps
	echo "setup done"

install:
	GO111MODULE=off go get -u github.com/golang/protobuf/protoc-gen-go
	GO111MODULE=off go get -u github.com/mwitkow/go-proto-validators/protoc-gen-govalidators
	GO111MODULE=off go get -u github.com/golang/mock/gomock
	GO111MODULE=on go get -u github.com/golang/mock/mockgen
	GO111MODULE=on go get -u google.golang.org/grpc

deps:
	go mod tidy

vendor:
	go mod vendor

gen:
	@make gen-proto
	@make gen-mock

gen-proto:
	#pbbase/*.proto
	mkdir -p ./internal/common/grpc/pbbase
	protoc -I/usr/local/include -I./api/p2pderivatives-proto -I${GOPATH}/src \
		--go_out=./internal/common/grpc/pbbase\
		./api/p2pderivatives-proto/method_option.proto
	#user/*.proto
	protoc -I./internal/user -I./api/p2pderivatives-proto -I${GOPATH}/src -I${GOPATH}/src/github.com/mwitkow/go-proto-validators -I/usr/local/include \
		--go_out=plugins=grpc:./internal/user/usercontroller user.proto \
		--govalidators_out=gogoimport=true:./internal/user/usercontroller user.proto
	#authentication/*.proto
	protoc -I./internal/authentication -I./api/p2pderivatives-proto -I${GOPATH}/src -I${GOPATH}/src/github.com/mwitkow/go-proto-validators -I/usr/local/include \
		--go_out=plugins=grpc:./internal/authentication authentication.proto \
		--govalidators_out=gogoimport=true:./internal/authentication authentication.proto
	#test/*.proto
	protoc -I./test -I./api/p2pderivatives-proto -I${GOPATH}/src -I${GOPATH}/src/github.com/mwitkow/go-proto-validators -I/usr/local/include \
		--go_out=plugins=grpc:./test test.proto

gen-mock:
	mockgen -destination test/mocks/mock_usercontroller/mock_controller.go  p2pderivatives-server/internal/user/usercontroller User_GetUserListServer
	mockgen -destination test/mocks/mock_usercommon/mock_service.go  p2pderivatives-server/internal/user/usercommon ServiceIf

client:
	mkdir -p bin
	go build -o ./bin/p2pdclient ./cmd/p2pdcli/p2pdcli.go

server:
	mkdir -p bin
	go build -o ./bin/server ./cmd/p2pdserver/server.go

bin:
	@make client
	@make server

run-server-local:
	@make server
	./bin/server -config ./test/config -appname p2pd -e integration -migrate

docker:
	docker build -t docker.pkg.github.com/cryptogarageinc/p2pderivatives-server/server .

run-docker:
	docker run -p 8080:8080 docker.pkg.github.com/cryptogarageinc/p2pderivatives-server/server

test-local:
	go test ./...

help:
	@make2help $(MAKEFILE_LIST)
