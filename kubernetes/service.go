package kubernetes

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/kubeconfig"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	Config *config.KubernetesConfig
}

func NewFromConfig(cf *config.KubernetesConfig) *Service {
	return &Service{cf}
}

func (s *Service) CreateClientset() (*kubernetes.Clientset, error) {
	config, err := kubeconfig.Load(s.Config.Context, s.Config.ConfigFile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
