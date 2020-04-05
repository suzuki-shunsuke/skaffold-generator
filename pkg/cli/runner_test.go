package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigParser_SetManifests(t *testing.T) {
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
	parser := &ConfigParser{}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := parser.SetManifests(d.cfg, d.manifests)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, d.cfg)
		})
	}
}

func TestConfigParser_SetArtifacts(t *testing.T) {
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
	parser := &ConfigParser{}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			err := parser.SetArtifacts(d.cfg, d.artifacts)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, d.cfg)
		})
	}
}

func TestConfigParser_Parse(t *testing.T) {
	data := []struct {
		title   string
		cfg     *Config
		targets map[string]struct{}
		exp     map[string]interface{}
		isErr   bool
	}{
		{
			title: "normal",
			cfg: &Config{
				Services: []ServiceConfig{
					{
						Name:      "mongodb",
						Manifests: []string{"mongodb.yaml"},
						Artifacts: []interface{}{
							map[string]interface{}{
								"image": "nginx",
							},
						},
					},
				},
			},
			targets: map[string]struct{}{"mongodb": struct{}{}},
			exp: map[string]interface{}{
				"build": map[string]interface{}{
					"artifacts": []interface{}{
						map[string]interface{}{
							"image": "nginx",
						},
					},
				},
				"deploy": map[string]interface{}{
					"kubectl": map[string]interface{}{
						"manifests": []string{"mongodb.yaml"},
					},
				},
			},
		},
	}
	parser := &ConfigParser{}
	for _, d := range data {
		t.Run(d.title, func(t *testing.T) {
			cfg, err := parser.Parse(d.cfg, d.targets)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, cfg)
		})
	}
}
