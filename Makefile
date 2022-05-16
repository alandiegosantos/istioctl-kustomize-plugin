IMG ?= alandiegosantos/istioctl-kustomize-plugin:0.0.1

test: main.go deps.go go.sum go.mod
	cat test/istio-operator.yaml | go run main.go -

Dockerfile: main.go deps.go go.sum go.mod
	go run main.go gen .
	sed -i.bak 's/golang:1.16-alpine/golang:1.18-alpine/g' Dockerfile && rm Dockerfile.bak
	sed -i.bak 's|RUN|RUN --mount=type=cache,sharing=locked,target=/root/.cache/go-build --mount=type=cache,sharing=locked,target=/gopath/pkg/mod|g' Dockerfile && rm Dockerfile.bak

.PHONY: image
image: Dockerfile
	docker build . -t ${IMG}

e2e: image
	kustomize build --enable-alpha-plugins test

go-test: main.go main_test.go go.sum go.mod
	go test ./...
