.PHONY: assemble_proto

asseble_proto: /usr/bin/sudo PATH=$GOPATH/bin /usr/local/bin/protoc --proto_path=protoc\
	--go_out=protoc --go_opt=paths=source_relative --go-grpc_out=protoc\
	--go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional protoc/*.proto

start_evans: PATH=$GOPATH/bin evans --proto protoc/chat.proto repl -p 8081

install_package: PATH=$GOPATH/bin /usr/local/go/bin/go install github.com/ktr0731/evans@v0.10.11
