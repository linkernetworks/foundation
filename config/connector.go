package config

import (
	"k8s.io/client-go/kubernetes"
)

type KubernetesConnector interface {
	Connect(clientset *kubernetes.Clientset)
}
