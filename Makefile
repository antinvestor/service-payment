# Service-specific configuration
SERVICE_NAME := payment
APP_DIRS     := apps/default apps/billing apps/checkout apps/ledger apps/integrations/airtel apps/integrations/jenga-api apps/integrations/mpesa apps/integrations/mtn apps/integrations/pawapay apps/integrations/polar apps/integrations/stripe apps/integrations/flutterwave apps/integrations/yellowcard

# Bootstrap: download shared Makefile.common if missing
ifeq (,$(wildcard .tmp/Makefile.common))
  $(shell mkdir -p .tmp && curl -sSfL https://raw.githubusercontent.com/antinvestor/common/main/Makefile.common -o .tmp/Makefile.common)
endif

include .tmp/Makefile.common

# Dart proto modules — each gets its own buf.gen.dart.<module>.yaml so that
# generation is scoped to a single module and dart packages don't leak each
# other's types. See proto/buf.gen.dart.*.yaml.
DART_MODULES := payment ledger billing

.PHONY: proto-generate-dart
proto-generate-dart: $(BIN)/buf ## Regenerate the per-module dart SDKs
	@if [ ! -d "$(PROTO_DIR)" ]; then exit 0; fi
	@# Wipe each module's lib/src/v1/ before regen so cross-module files
	@# left behind by older unscoped runs are removed. Each module's
	@# own .pb.dart is then re-emitted by the per-module buf generate
	@# below. (Proto paths here are flat — proto/<m>/v1/<m>.proto —
	@# so all three dart packages place their files in lib/src/v1/.)
	@for m in $(DART_MODULES); do \
		rm -rf sdk/dart/$$m/lib/src/v1; \
	done
	@for m in $(DART_MODULES); do \
		echo "==> dart $$m"; \
		(cd $(PROTO_DIR) && buf generate --template buf.gen.dart.$$m.yaml $$m); \
	done

.PHONY: proto-generate-checkout
proto-generate-checkout: $(BIN)/buf ## Regenerate checkout Go stubs locally
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	@(cd $(PROTO_DIR) && PATH="$$(go env GOPATH)/bin:$$PATH" buf generate --template buf.gen.checkout.yaml --path payment/checkout)

.PHONY: proto-generate-billing-collection
proto-generate-billing-collection: $(BIN)/buf ## Regenerate billing collection Go stubs locally
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	@(cd $(PROTO_DIR) && PATH="$$(go env GOPATH)/bin:$$PATH" buf generate --template buf.gen.billing.collection.yaml --path billing/collection)

# Wire dart generation into the standard proto-generate pipeline.
proto-generate: proto-generate-dart proto-generate-checkout proto-generate-billing-collection

format: ## Format Go files (used by pre-commit hook)
	gofmt -w .
