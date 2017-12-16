package config

import (
	"bitbucket.org/linkernetworks/aurora/src/kubeconfig"
	"k8s.io/client-go/kubernetes"
)

type DeploymentConfig struct {
	Type       string               `json:"type"`
	Kubernetes KubeDeploymentConfig `json:"kubernetes"`
}

type KubeDeploymentConfig struct {
	ConfigFile string `json:"config"`
	Context    string `json:"context"`
	Namespace  string `json:"namespace"`
}

func (kcf KubeDeploymentConfig) CreateClientset() (*kubernetes.Clientset, error) {
	config, err := kubeconfig.Load("", kcf.ConfigFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
