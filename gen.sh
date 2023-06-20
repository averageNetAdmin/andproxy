export GOPATH=$HOME/go
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin
protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models access_control.proto
protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models server.proto
# protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models firewall.proto
protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models cache.proto
protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models http_handler.proto
protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models l4_handler.proto
protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models handler.proto
protoc -I=andproto/models --go_opt=paths=source_relative --go_out=andproto/models stats.proto
protoc -I=andproto -I=andproto/models --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=andproto --go-grpc_out=andproto dbmng.proto
protoc -I=andproto -I=andproto/models --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go_out=andproto --go-grpc_out=andproto proxy.proto
