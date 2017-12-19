package config

type KubernetesConfig struct {
	ConfigFile string `json:"config"`
	Context    string `json:"context"`
	Namespace  string `json:"namespace"`
	InCluster  bool   `json:"inCluster"`
}
