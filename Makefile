API_PATH=api/p2pderivatives-proto

setup: install install-tools gen deps
	@echo "setup done"

install: 
	go mod download

install-tools:
	$(shell cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %)

deps:
	go mod tidy

vendor:
	go mod vendor

gen: gen-proto gen-mock gen-ssl-certs

gen-proto:
	#pbbase/*.proto
	mkdir -p ./internal/common/grpc/pbbase
	$(call gen_proto_go,${API_PATH}, method_option)
	#user/*.proto
	$(call gen_proto_go,${API_PATH}, user)
	#authentication/*.proto
	$(call gen_proto_go,${API_PATH}, authentication)
	#test/*.proto
	$(call gen_proto_go,test, test)

define gen_proto_go
	protoc --proto_path=./$1 -I./api/p2pderivatives-proto  --go_out=plugins=grpc:../ --govalidators_out=../ $2.proto
endef

gen-mock:
	mkdir -p test/mocks/mock_usercontroller
	mockgen -destination test/mocks/mock_usercontroller/mock_controller.go  p2pderivatives-server/internal/user/usercontroller User_GetUserListServer,User_ReceiveDlcMessagesServer,User_GetConnectedUsersServer
	mkdir -p test/mocks/mock_usercommon
	mockgen -destination test/mocks/mock_usercommon/mock_service.go  p2pderivatives-server/internal/user/usercommon ServiceIf

gen-ssl-certs:
	mkdir -p certs/db
	$(eval CERT_TEMP=$(shell mktemp -d))
	openssl req -new -text -passout pass:abcd -subj /CN=localhost -out ${CERT_TEMP}/db.req -keyout ${CERT_TEMP}/privkey.pem
	openssl rsa -in ${CERT_TEMP}/privkey.pem -passin pass:abcd -out certs/db.key
	openssl req -x509 -in ${CERT_TEMP}/db.req -text -key certs/db.key -out certs/db.crt
	chmod 600 certs/db.key

client:
	mkdir -p bin
	go build -o ./bin/p2pdclient ./cmd/p2pdcli/p2pdcli.go

server:
	mkdir -p bin
	go build -o ./bin/server ./cmd/p2pdserver/server.go

bin: client server

run-db:
	docker-compose up db

run-server-local: server
	./bin/server -config ./test/config -appname p2pd -e integration -migrate

docker:
	docker build -t docker.pkg.github.com/cryptogarageinc/p2pderivatives-server/server .

run-docker:
	docker-compose up

test-local:
	go test ./...

help:
	@make2help $(MAKEFILE_LIST)
