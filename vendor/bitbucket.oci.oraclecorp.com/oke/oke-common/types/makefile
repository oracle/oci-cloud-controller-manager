
.DEFAULT_GOAL := protobuf

# Ordering of .proto files matters - inserting a new file in the middle
# of the list of .proto files will rebuild all files following it and
# potentially wreak havoc by causing incompatibility.	 Please add new
# .proto files only to the end of the list of .proto files.
.PHONY: protobuf
protobuf:
	protoc -I . -I ../proto -I ../vendor --go_out=plugins=grpc:$$GOPATH/src \
		../proto/k8instance.proto \
		../proto/bmcs.proto \
		../proto/tenantagent.proto \
		../proto/error.proto \
		../proto/tmcollector.proto \
		../proto/listoptions.proto \
		../proto/settings.proto \
		../proto/networkconfigs.proto \
		../proto/resourceprincipals.proto

