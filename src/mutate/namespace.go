package mutate

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func getFromNamespace(namespaceName string) (string, error) {
	if len(namespaceName) == 0 {
		return "", nil
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}
	namespace, err := clientset.CoreV1().Namespaces().Get(context.Background(), namespaceName, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	if authpath, ok := namespace.Annotations["govaultenv.io/authpath"]; ok {
		return authpath, nil
	}
	return "", nil
}
