package config

/*
import (
	"bitbucket.org/linkernetworks/aurora/src/service/kubernetes"
	"errors"

	jsonpatch "github.com/evanphx/json-patch"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func LoadRemoteService(c Config) (Config, error) {
	var dst = c

	if dst.KubernetesConfig == nil {
		return dst, errors.New("kubernetes config is not defined, can't convert config to load kubernetes service")
	}

	k8s := kubernetes.NewFromConfig(c.Kubernetes)
	clientset := k8s.CreateClientset()

	res, err := clientset.Core()
		.Pods("default")
		.Patch("mongo-0", []byte{`[ { "op": "add", "path": "/metadata/labels/podindex", "value": "0" } ]`})
	_ = res
	if err != nil {
		return dst, err
	}

	return dst, nil
}

func NewExternalMongoService() v1.Service {
	return v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "mongo-external",
			Labels: map[string]string{
				"role": "mongo",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				"role":     "mongo",
				"podindex": "0",
			},
			Ports: []v1.ServicePort{
				{
					Port:       27017,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: 27017},
					NodePort:   31717,
				},
			},
		},
	}
}
*/
