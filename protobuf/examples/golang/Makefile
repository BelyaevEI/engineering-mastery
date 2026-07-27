.PHONY: generate clean deps contracts

# Пути
PROTO_DIR = api
PKG_DIR = pkg
DOC_DIR = doc

generate: deps
	@echo "Генерация кода из .proto файлов..."
	@protoc \
		-I$(PROTO_DIR) \
		-Ithird_party \
		-Ithird_party/grpc-gateway \
		--go_out=$(PKG_DIR) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(PKG_DIR) \
		--go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PKG_DIR) \
		--grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(DOC_DIR) \
		--openapiv2_opt=allow_merge=true \
		--openapiv2_opt=json_names_for_fields=false \
		--openapiv2_opt=merge_file_name=user_service \
		--validate_out="lang=go:$(PKG_DIR)" \
		--validate_opt=paths=source_relative \
		$(PROTO_DIR)/user_service.proto

	@echo "Генерация завершена!"

# Генерация контрактов из proto (Swagger/OpenAPI спецификации)
contracts: deps
	@echo "Генерация контрактов из .proto файлов..."
	@protoc \
		-I$(PROTO_DIR) \
		-I$(THIRD_PARTY_DIR) \
		-I$(THIRD_PARTY_DIR)/grpc-gateway \
		--openapiv2_out=$(DOC_DIR) \
		--openapiv2_opt=allow_merge=true \
		--openapiv2_opt=json_names_for_fields=false \
		--openapiv2_opt=merge_file_name=user_service \
		$(PROTO_DIR)/user_service.proto

	@echo "Контракты сгенерированы в $(DOC_DIR)/"

deps:
	@echo "Установка зависимостей protoc plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@go install github.com/envoyproxy/protoc-gen-validate@latest

clean:
	@echo "Очистка сгенерированных файлов..."
	@rm -f $(PKG_DIR)/user_service.pb.go
	@rm -f $(PKG_DIR)/user_service.pb.gw.go
	@rm -f $(PKG_DIR)/user_service_grpc.pb.go
	@rm -f $(PKG_DIR)/user_service.pb.validate.go
	@rm -f $(DOC_DIR)/user_service.swagger.json
