
.DEFAULT_GOAL := protobuf

# Ordering of .proto files matters - inserting a new file in the middle
# of the list of .proto files will rebuild all files following it and
# potentially wreak havoc by causing incompatibility.  Please add new
# .proto files only to the end of the list of .proto files.
.PHONY: protobuf
protobuf:
	protoc -I . -I ../../proto -I ../../vendor --go_out=plugins=grpc:$$GOPATH/src \
		../../proto/nodepools/nodepools.proto
