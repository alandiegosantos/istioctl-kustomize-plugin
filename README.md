# Istio Operator kustomize plugin

Istio Operator plugin to kustomize to generate the manifests using istioctl plugin. 

# The why

While maintaining multiple services in kubernetes, we usually rely on kustomize to manage and patch the manifests, so kustomize is widely used in pipelines. To install Istio, we rely on the Istio Operator resource to configure Istio resources not needing to manage multiple manifests. 
So, the main purpose of this plugin is to enable kustomize to generate the manifests to install and configure Istio.

# How to build it

To filter consists of a Docker image. To build the image:
```
$ make image
```

# How to use it
To generate the manifests, made the necessary changes in the ./test/istio-operator.yaml manifest, and run:
```
$ kustomize build --enable-alpha-plugins test
```

# How to improve
Right now, the plugin is configured with Istio 1.13 manifests. To change the version:
```
$ go get istio.io/istio@<branch with the version on github.com/istio/istio>
$ go mod tidy
```

Example for the istioctl 1.13:
```
$ go get istio.io/istio@release-1.13
```