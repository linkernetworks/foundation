package config

type DeploymentConfig struct {
	Type       string               `json:"type"`
	Kubernetes KubeDeploymentConfig `json:"kubernetes"`
}

type KubeDeploymentConfig struct {
	ConfigFile string `json:"config"`
	Context    string `json:"context"`
	Namespace  string `json:"namespace"`
}
