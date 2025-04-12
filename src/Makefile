PROTO_DIR = protofiles
PROTO_OUT_DIR = .

PROTOC_FLAGS = --go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
           --go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative

proto:
	protoc $(PROTOC_FLAGS) $(PROTO_DIR)/*.proto

clean:
	rm -f $(PROTO_DIR)/*.pb.go