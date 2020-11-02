package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"testing"
)

func TestImageScan(t *testing.T) {
	obj, err := yaml.Parse(`
version: '3'
services:
  podinfo:
    build: ./podinfo
    image: chanwit/podinfo:v5.0.3
  nginx:
    image: nginx:latest
`)
	assert.NoError(t, err)
	services, err := obj.Pipe(yaml.Lookup("services"))
	services.VisitFields(func(node *yaml.MapNode) error {
		image := node.Value.Field("image")
		if image != nil {
			fmt.Print(yaml.GetValue(image.Value))
		}
		return nil
	})
	assert.NoError(t, err)
}

func TestServiceScan(t *testing.T) {
	obj, err := yaml.Parse(`
version: '3'
services:
  podinfo:
    build: ./podinfo
    image: chanwit/podinfo:v5.0.3
  nginx:
    image: nginx:latest
`)
	assert.NoError(t, err)
	services, err := obj.Pipe(yaml.Lookup("services"))
	services.VisitFields(func(node *yaml.MapNode) error {
		service := yaml.GetValue(node.Key)
		fmt.Println(service)
		return nil
	})
	assert.NoError(t, err)
}
