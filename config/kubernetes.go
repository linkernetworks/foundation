package config

type KubernetesConfig struct {
	ConfigFile string `json:"config"`
	Context    string `json:"context"`
	Namespace  string `json:"namespace"`
	InCluster  bool   `json:"inCluster"`
	OutCluster *KubernetesOutClusterConfig
}

func (c *KubernetesConfig) LoadDefaults() error {
	if c.Namespace == "" {
		c.Namespace = "default"
	}
	if c.OutCluster == nil {
		c.OutCluster = &KubernetesOutClusterConfig{}
		LoadDefaults(c.OutCluster)
	}
	return nil
}

type KubernetesOutClusterConfig struct {
	AddressType string `json:"addressType"`
}

func (c *KubernetesOutClusterConfig) LoadDefaults() error {
	if c.AddressType == "" {
		c.AddressType = "ExternalIP"
	}
	return nil
}
