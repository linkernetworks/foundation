package kubernetes

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/kubeconfig"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type Service struct {
	Config *config.KubernetesConfig
}

func NewFromConfig(cf *config.KubernetesConfig) *Service {
	return &Service{cf}
}

// Create the kubernetes config
func (s *Service) LoadConfig() (*rest.Config, error) {
	if s.Config.InCluster {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	config, err := kubeconfig.Load(s.Config.Context, s.Config.ConfigFile)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *Service) CreateClientset() (*kubernetes.Clientset, error) {
	config, err := s.LoadConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
