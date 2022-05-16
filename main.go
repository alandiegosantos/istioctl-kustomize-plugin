package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/yaml"

	"istio.io/istio/operator/pkg/helm"
	"istio.io/istio/operator/pkg/manifest"
	"istio.io/istio/operator/pkg/name"
	"istio.io/istio/operator/pkg/object"
	"istio.io/istio/operator/pkg/util/clog"
)

const (
	istioOperatorKind       = "IstioOperator"
	istioOperatorApiVersion = "install.istio.io/v1alpha1"
)

type IstioOperator struct {
	Value string `yaml:"value" json:"value"`
}

func init() {
	log.SetOutput(ioutil.Discard)
}

func main() {
	fn := func(items []*kyaml.RNode) ([]*kyaml.RNode, error) {

		result := []*kyaml.RNode{}

		for _, item := range items {

			if item.GetApiVersion() == istioOperatorApiVersion && item.GetKind() == istioOperatorKind {
				manifests, err := generateIstioManifests(item)
				if err != nil {
					return nil, err
				}
				result = append(result, manifests...)

			} else {
				// Just append the non IstioOperator manifests
				if err := item.PipeE(kyaml.SetAnnotation("custom.io/the-value", "some-value")); err != nil {
					return nil, err
				}

				result = append(result, item)
			}

		}

		return result, nil
	}
	p := framework.SimpleProcessor{Filter: kio.FilterFunc(fn)}
	cmd := command.Build(p, command.StandaloneDisabled, false)
	command.AddGenerateDockerfile(cmd)
	if err := cmd.Execute(); err != nil {
		fmt.Printf("Error while generating manifests: %v", err)
		os.Exit(1)
	}
}

// Copied from istioctl code
// https://github.com/istio/istio/blob/764e6688921e326bdaa530d5ef3a9ed5d87372f7/operator/cmd/mesh/manifest-generate.go#L159
// orderedManifests generates a list of manifests from the given map sorted by the default object order
// This allows
func orderedManifests(mm name.ManifestMap) ([]string, error) {
	var rawOutput []string
	var output []string
	for _, mfs := range mm {
		rawOutput = append(rawOutput, mfs...)
	}
	objects, err := object.ParseK8sObjectsFromYAMLManifest(strings.Join(rawOutput, helm.YAMLSeparator))
	if err != nil {
		return nil, err
	}
	// For a given group of objects, sort in order to avoid missing dependencies, such as creating CRDs first
	objects.Sort(object.DefaultObjectOrder())
	for _, obj := range objects {
		yml, err := obj.YAML()
		if err != nil {
			return nil, err
		}
		output = append(output, string(yml))
	}

	return output, nil
}

func generateIstioManifests(item *kyaml.RNode) ([]*kyaml.RNode, error) {

	result := []*kyaml.RNode{}

	file, err := ioutil.TempFile("/tmp", "istio-operator.*.yaml")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create temporary file")
	}
	defer os.Remove(file.Name())

	itemJSON, err := item.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal item")
	}

	itemYAML, err := yaml.JSONToYAML(itemJSON)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse item")
	}

	_, err = file.Write(itemYAML)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write file")
	}

	if err = file.Close(); err != nil {
		return nil, errors.Wrap(err, "failed to close file")
	}

	l := clog.NewConsoleLogger(ioutil.Discard, ioutil.Discard, nil)
	manifests, _, err := manifest.GenManifests([]string{file.Name()}, nil, false, nil, l)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render the manifests")
	}

	ordered, err := orderedManifests(manifests)
	if err != nil {
		return nil, errors.Wrap(err, "failed to order manifests")
	}

	for _, m := range ordered {
		n, err := kyaml.Parse(m)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse manifest")
		}
		result = append(result, n)
	}

	return result, nil

}
