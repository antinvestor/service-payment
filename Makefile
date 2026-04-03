# Service-specific configuration
SERVICE_NAME := payment
APP_DIRS     := apps/default apps/billing apps/ledger apps/integrations/airtel apps/integrations/jenga-api apps/integrations/mpesa apps/integrations/mtn apps/integrations/polar apps/integrations/stripe

# Bootstrap: download shared Makefile.common if missing
ifeq (,$(wildcard .tmp/Makefile.common))
  $(shell mkdir -p .tmp && curl -sSfL https://raw.githubusercontent.com/antinvestor/common/main/Makefile.common -o .tmp/Makefile.common)
endif

include .tmp/Makefile.common

format: ## Format Go files (used by pre-commit hook)
	gofmt -w .
