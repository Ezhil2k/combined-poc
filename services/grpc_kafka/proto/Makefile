PROTOC_GEN_GO := protoc-gen-go
PROTOC_GEN_GO_GRPC := protoc-gen-go-grpc
PROTOC := protoc

PROTO_DIR := .
OUT_DIR := ./trade_engine_walletpb

all: go

go:
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUT_DIR) \
		--go-grpc_out=$(OUT_DIR) \
		wallet.proto

clean:
	rm -f $(OUT_DIR)/*.go 