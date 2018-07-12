package config

type DeploymentConfig struct {
	Type       string           `json:"type"`
	Kubernetes KubernetesConfig `json:"kubernetes"`
}
