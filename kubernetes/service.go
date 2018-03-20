package kubernetes

import (
	"bitbucket.org/linkernetworks/aurora/src/config"
	"bitbucket.org/linkernetworks/aurora/src/kubeconfig"
	"bitbucket.org/linkernetworks/aurora/src/logger"
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

	logger.Debugf("Loading kubernetes config with context=%s from config=%s", s.Config.Context, s.Config.ConfigFile)
	config, err := kubeconfig.Load(s.Config.Context, s.Config.ConfigFile)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (s *Service) NewClientset() (*kubernetes.Clientset, error) {
	config, err := s.LoadConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
