BINARY_NAME ?= scip-search
BUILD_DIR ?= build
GO ?= go
INSTALL_DIR ?= $(HOME)/.local/bin
SOURCE_REF ?= local
SOURCE_REVISION ?= unknown

VERSION_LDFLAGS := \
	-X 'scip-search/internal/version.SourceRef=$(SOURCE_REF)' \
	-X 'scip-search/internal/version.SourceRevision=$(SOURCE_REVISION)'

.PHONY: build check-testhelpers install sync-embedded test validate-distribution

test:
	go test ./...

validate-distribution:
	@GO="$(GO)" ./scripts/validate-distribution.sh

build:
	@command -v "$(GO)" >/dev/null 2>&1 || { printf 'Go is required for source builds: %s\n' "$(GO)" >&2; exit 1; }
	@mkdir -p "$(BUILD_DIR)"
	@$(GO) build \
		-ldflags "$(VERSION_LDFLAGS)" \
		-o "$(BUILD_DIR)/$(BINARY_NAME)" \
		./cmd/scip-search || { printf 'source build failed\n' >&2; exit 1; }

install: build
	@dest="$(INSTALL_DIR)/$(BINARY_NAME)"; \
	mkdir -p "$(INSTALL_DIR)" || { printf 'INSTALL_DIR is not usable: %s\n' "$(INSTALL_DIR)" >&2; exit 1; }; \
	cp "$(BUILD_DIR)/$(BINARY_NAME)" "$$dest" || { printf 'INSTALL_DIR is not usable: %s\n' "$(INSTALL_DIR)" >&2; exit 1; }; \
	chmod 0755 "$$dest" || { printf 'INSTALL_DIR is not usable: %s\n' "$(INSTALL_DIR)" >&2; exit 1; }; \
	if [ ! -x "$$dest" ]; then \
		printf 'installed scip-search is not executable: %s\n' "$$dest" >&2; \
		exit 1; \
	fi; \
	if ! version_output=$$("$$dest" --version 2>&1); then \
		printf 'installed scip-search at %s failed --version\n' "$$dest" >&2; \
		exit 1; \
	fi; \
	case "$$version_output" in \
	*"source"* ) ;; \
	*) printf 'installed scip-search at %s did not report source provenance\n' "$$dest" >&2; exit 1 ;; \
	esac; \
	case "$$version_output" in \
	*"$(SOURCE_REF)"*"$(SOURCE_REVISION)"* ) ;; \
	*) printf 'installed scip-search at %s did not report requested source provenance\n' "$$dest" >&2; exit 1 ;; \
	esac; \
	printf 'Installed scip-search source ref=%s revision=%s to %s\n' "$(SOURCE_REF)" "$(SOURCE_REVISION)" "$$dest"

sync-embedded:

check-testhelpers:
