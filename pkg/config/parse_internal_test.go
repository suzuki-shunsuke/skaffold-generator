package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_calcTargets(t *testing.T) {
	data := []struct {
		title   string
		cfg     map[string]map[string]struct{}
		targets map[string]struct{}
		exp     map[string]struct{}
		isErr   bool
	}{
		{
			title: "normal",
			cfg: map[string]map[string]struct{}{
				"mongodb": {},
				"app":     {},
			},
			targets: map[string]struct{}{"mongodb": {}},
			exp:     map[string]struct{}{"mongodb": {}},
		},
		{
			title: "dependency",
			cfg: map[string]map[string]struct{}{
				"mongodb": {},
				"app": {
					"mongodb": {},
				},
				"foo": {},
			},
			targets: map[string]struct{}{"app": {}},
			exp: map[string]struct{}{
				"app":     {},
				"mongodb": {},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			targets := map[string]struct{}{}
			err := calcTargets(d.cfg, d.targets, targets)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, targets)
		})
	}
}

func Test_calcTargetsRecursively(t *testing.T) { //nolint:funlen
	data := []struct {
		title   string
		cfg     map[string]map[string]struct{}
		target  string
		targets map[string]struct{}
		exp     map[string]struct{}
		isErr   bool
	}{
		{
			title: "normal",
			cfg: map[string]map[string]struct{}{
				"mongodb": {},
				"app":     {},
			},
			target:  "mongodb",
			targets: map[string]struct{}{},
			exp:     map[string]struct{}{"mongodb": {}},
		},
		{
			title: "dependency",
			cfg: map[string]map[string]struct{}{
				"mongodb": {},
				"app": {
					"mongodb": {},
				},
			},
			target:  "app",
			targets: map[string]struct{}{},
			exp: map[string]struct{}{
				"app":     {},
				"mongodb": {},
			},
		},
		{
			title: "dependency 2",
			cfg: map[string]map[string]struct{}{
				"mongodb": {},
				"app": {
					"mongodb": {},
					"api":     {},
				},
				"api": {
					"mongodb": {},
				},
				"foo": {},
				"bar": {},
			},
			target:  "app",
			targets: map[string]struct{}{"foo": {}},
			exp: map[string]struct{}{
				"app":     {},
				"api":     {},
				"mongodb": {},
				"foo":     {},
			},
		},
		{
			title: "circular dependency",
			cfg: map[string]map[string]struct{}{
				"mongodb": {
					"app": {},
				},
				"app": {
					"mongodb": {},
				},
			},
			target:  "app",
			targets: map[string]struct{}{},
			exp: map[string]struct{}{
				"app":     {},
				"mongodb": {},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			err := calcTargetsRecursively(d.cfg, d.target, d.targets)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, d.targets)
		})
	}
}

func Test_setArtifacts(t *testing.T) {
	data := []struct {
		title     string
		cfg       map[string]interface{}
		artifacts []interface{}
		exp       map[string]interface{}
		isErr     bool
	}{
		{
			title: "base is not found",
			cfg:   map[string]interface{}{},
			artifacts: []interface{}{
				map[string]interface{}{
					"image": "nginx",
				},
			},
			exp: map[string]interface{}{
				"build": map[string]interface{}{
					"artifacts": []interface{}{
						map[string]interface{}{
							"image": "nginx",
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			err := setArtifacts(d.cfg, d.artifacts)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, d.cfg)
		})
	}
}

func Test_setManifests(t *testing.T) {
	data := []struct {
		title     string
		cfg       map[string]interface{}
		manifests []string
		exp       map[string]interface{}
		isErr     bool
	}{
		{
			title:     "base is not found",
			cfg:       map[string]interface{}{},
			manifests: []string{"deployment.yaml"},
			exp: map[string]interface{}{
				"deploy": map[string]interface{}{
					"kubectl": map[string]interface{}{
						"manifests": []string{"deployment.yaml"},
					},
				},
			},
		},
		{
			title: "base is merged",
			cfg: map[string]interface{}{
				"apiVersion": "skaffold/v2beta1",
				"deploy": map[string]interface{}{
					"kubectl": map[string]interface{}{
						"manifests": []string{"deployment.yaml"},
						"flags": map[string]interface{}{
							"disableValidation": false,
						},
					},
				},
			},
			manifests: []string{"service.yaml"},
			exp: map[string]interface{}{
				"apiVersion": "skaffold/v2beta1",
				"deploy": map[string]interface{}{
					"kubectl": map[string]interface{}{
						"manifests": []string{"service.yaml"},
						"flags": map[string]interface{}{
							"disableValidation": false,
						},
					},
				},
			},
		},
	}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			err := setManifests(d.cfg, d.manifests)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, d.cfg)
		})
	}
}
