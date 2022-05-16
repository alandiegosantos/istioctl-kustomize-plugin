package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestIstioManifests(t *testing.T) {

	testFile := "./test/istio-operator.yaml"

	item, err := kyaml.ReadFile(testFile)
	assert.Nil(t, err, "failed to open test file")

	json, err := item.MarshalJSON()

	fmt.Printf("item %v\n", string(json))

	result, err := generateIstioManifests(item)
	assert.Nil(t, err, "generateIstioManifests returned an error")
	assert.Greater(t, len(result), 0, "generateIstioManifests returned an empty result")

	for _, m := range result {
		_, err := m.MarshalJSON()
		assert.Nil(t, err, "marshalling to JSON", m.GetKind())
	}

}
