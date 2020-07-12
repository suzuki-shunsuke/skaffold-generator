package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigParser_calcTargets(t *testing.T) { //nolint:funlen
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
	parser := &ConfigParser{}
	for _, d := range data {
		d := d
		t.Run(d.title, func(t *testing.T) {
			err := parser.calcTargets(d.cfg, d.target, d.targets)
			if d.isErr {
				require.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			require.Equal(t, d.exp, d.targets)
		})
	}
}
