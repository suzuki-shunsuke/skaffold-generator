package config_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/config"
)

func TestParser_Parse(t *testing.T) { //nolint:funlen
	data := []struct {
		title   string
		cfg     config.Config
		targets map[string]struct{}
		exp     map[string]interface{}
		isErr   bool
	}{
		{
			title: "normal",
			cfg: config.Config{
				Services: []config.ServiceConfig{
					{
						Name:      "mongodb",
						Manifests: []string{"mongodb.yaml"},
						Artifacts: []interface{}{
							map[string]interface{}{
								"image": "mongodb",
							},
						},
					},
				},
			},
			targets: map[string]struct{}{"mongodb": {}},
			exp: map[string]interface{}{
				"build": map[string]interface{}{
					"artifacts": []interface{}{
						map[string]interface{}{
							"image": "mongodb",
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
		{
			title: "dependency",
			cfg: config.Config{
				Services: []config.ServiceConfig{
					{
						Name:      "mongodb",
						Manifests: []string{"mongodb.yaml"},
						Artifacts: []interface{}{
							map[string]interface{}{
								"image": "mongodb",
							},
						},
					},
					{
						Name:      "app",
						Manifests: []string{"app.yaml"},
						Artifacts: []interface{}{
							map[string]interface{}{
								"image": "nginx",
							},
						},
						DependsOn: []string{"mongodb"},
					},
					{
						Name:      "foo",
						Manifests: []string{"foo.yaml"},
						Artifacts: []interface{}{
							map[string]interface{}{
								"image": "foo",
							},
						},
					},
				},
			},
			targets: map[string]struct{}{"app": {}},
			exp: map[string]interface{}{
				"build": map[string]interface{}{
					"artifacts": []interface{}{
						map[string]interface{}{
							"image": "mongodb",
						},
						map[string]interface{}{
							"image": "nginx",
						},
					},
				},
				"deploy": map[string]interface{}{
					"kubectl": map[string]interface{}{
						"manifests": []string{"app.yaml", "mongodb.yaml"},
					},
				},
			},
		},
	}
	parser := config.Parser{}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			cfg, err := parser.Parse(d.cfg, d.targets)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			artifacts := cfg["build"].(map[string]interface{})["artifacts"].([]interface{})
			sort.Slice(artifacts, func(i, j int) bool {
				a := artifacts[i].(map[string]interface{})["image"].(string)
				b := artifacts[j].(map[string]interface{})["image"].(string)
				return a < b
			})
			manifests := cfg["deploy"].(map[string]interface{})["kubectl"].(map[string]interface{})["manifests"].([]string)
			sort.Strings(manifests)
			require.Equal(t, d.exp, cfg)
		})
	}
}
