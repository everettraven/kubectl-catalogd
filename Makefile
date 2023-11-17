include .bingo/Variables.mk

export CATALOGD_VERSION ?= $(shell go list -mod=mod -m -f "{{.Version}}" github.com/operator-framework/catalogd)
export GIT_VERSION       ?= $(shell git describe --tags --always --dirty)
export VERSION_PKG       ?= $(shell go list -m)/internal/cli
export GO_BUILD_ASMFLAGS ?= all=-trimpath=${PWD}
export GO_BUILD_LDFLAGS  ?= -s -w -X "$(VERSION_PKG).version=$(GIT_VERSION)" -X "$(VERSION_PKG).catalogd_version=$(CATALOGD_VERSION)"
export GO_BUILD_GCFLAGS  ?= all=-trimpath=${PWD}
CERT_MGR_VERSION ?= v1.9.0

.PHONY: build
build:
	go build \
	-asmflags '$(GO_BUILD_ASMFLAGS)' \
	-ldflags '$(GO_BUILD_LDFLAGS)' \
	-gcflags '$(GO_BUILD_GCFLAGS)' \
	-o kubectl-catalogd main.go

UNIT_TEST_DIRS=$(shell go list ./... | grep -v /test/)
.PHONY: unit
unit:
	go test -v $(UNIT_TEST_DIRS) -coverprofile cover.out

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_ARGS)

.PHONY: e2e
e2e: build delete-cluster create-cluster cert-manager image-registry build-push-e2e-catalog catalogd test-e2e delete-cluster

KIND_CLUSTER_NAME ?= kubectl-catalogd-e2e
.PHONY: create-cluster
create-cluster: $(KIND)
	$(KIND) create cluster --name $(KIND_CLUSTER_NAME)

.PHONY: delete-cluster
delete-cluster: $(KIND)
	$(KIND) delete cluster --name $(KIND_CLUSTER_NAME)

.PHONY: cert-manager
cert-manager:
	kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${CERT_MGR_VERSION}/cert-manager.yaml"
	kubectl wait --for=condition=Available --namespace="cert-manager" "deployment/cert-manager-webhook" --timeout="60s"

.PHONY: catalogd
catalogd: 
	kubectl apply -f "https://github.com/operator-framework/catalogd/releases/download/${CATALOGD_VERSION}/catalogd.yaml"
	kubectl wait --for=condition=Available --namespace="catalogd-system" "deployment/catalogd-controller-manager" --timeout="60s"

.PHONY: test-e2e
test-e2e:
	go test -v ./test/e2e/...

E2E_REGISTRY_NAME=docker-registry
E2E_REGISTRY_NAMESPACE=kubectl-catalogd-e2e
image-registry: ## Setup in-cluster image registry
	./test/tools/image-registry.sh ${E2E_REGISTRY_NAMESPACE} ${E2E_REGISTRY_NAME}

export CATALOG_IMG=${E2E_REGISTRY_NAME}.${E2E_REGISTRY_NAMESPACE}.svc:5000/test-catalog:e2e
build-push-e2e-catalog: ## Build the testdata catalog used for e2e tests and push it to the image registry
	./test/tools/build-push-e2e-catalog.sh ${E2E_REGISTRY_NAMESPACE} ${CATALOG_IMG}

all: lint build test e2e

export ENABLE_RELEASE_PIPELINE ?= false
export GORELEASER_ARGS ?= --snapshot --clean

.PHONY: release
release: $(GORELEASER) #EXHELP Runs goreleaser for the operator-controller. By default, this will run only as a snapshot and will not publish any artifacts unless it is run with different arguments. To override the arguments, run with "GORELEASER_ARGS=...". When run as a github action from a tag, this target will publish a full release.
	$(GORELEASER) $(GORELEASER_ARGS)
