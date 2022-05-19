IMG ?= alandiegosantos/istioctl-kustomize-plugin:0.0.1

containerized: image


test: main.go deps.go go.sum go.mod
	cat test/istio-operator.yaml | go run main.go -

plugins/istioctl-generator: main.go deps.go go.sum go.mod
	go mod download
	go build -v -o $@

Dockerfile: main.go deps.go go.sum go.mod
	go run main.go gen .
	sed -i.bak 's/golang:1.16-alpine/golang:1.18-alpine/g' Dockerfile && rm Dockerfile.bak
	sed -i.bak 's|RUN|RUN --mount=type=cache,sharing=locked,target=/root/.cache/go-build --mount=type=cache,sharing=locked,target=/gopath/pkg/mod|g' Dockerfile && rm Dockerfile.bak

.PHONY: image
image: Dockerfile
	docker build . -t ${IMG}

e2e-container: image
	kustomize build --enable-alpha-plugins test/containerized_krm

e2e-exec: plugins/istioctl-generator
	kustomize build --enable-alpha-plugins --enable-exec test/exec_krm

go-test: main.go main_test.go go.sum go.mod
	go test ./...
