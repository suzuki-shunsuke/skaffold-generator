package cli_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suzuki-shunsuke/skaffold-generator/pkg/cli"
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
	parser := cli.ConfigParser{}
	for _, d := range data {
		d := d
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
	parser := cli.ConfigParser{}
	for _, d := range data {
		d := d
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

func TestConfigParser_Parse(t *testing.T) { //nolint:funlen
	data := []struct {
		title   string
		cfg     *cli.Config
		targets map[string]struct{}
		exp     map[string]interface{}
		isErr   bool
	}{
		{
			title: "normal",
			cfg: &cli.Config{
				Services: []cli.ServiceConfig{
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
			cfg: &cli.Config{
				Services: []cli.ServiceConfig{
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
	parser := cli.ConfigParser{}
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

func TestConfigParser_CalcTargets(t *testing.T) {
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
	parser := cli.ConfigParser{}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			targets, err := parser.CalcTargets(d.cfg, d.targets)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, targets)
		})
	}
}
