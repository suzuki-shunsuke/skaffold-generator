package config

// Config is configuration of skaffold-generator (skaffold-generator.yaml).
type Config struct {
	// Services is a list of service configuration
	Services []ServiceConfig
	// Base is a base of generated `skaffold.yaml`.
	// `deploy.kubectl.manifests` and `build.artifacts` are overwritten.
	Base map[string]interface{}
}

type ServiceConfig struct {
	// Name is a service name.
	// Name must be unique.
	Name string
	// Artifacts is skaffold.yaml's `build.artifacts`
	Artifacts []interface{}
	// Manifests is skaffold.yaml's `deploy.kubectl.manifests`
	Manifests []string
	// DependsOn is service names which this service depends on
	DependsOn []string `yaml:"depends_on"`
}
