package kubernetes

import (
	"context"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListAllIngresses list all the ingresses from the specified namespace
// This paginates and returns all available ingresses from the cluster.
func (cfg *Config) ListAllIngresses() ([]netv1.Ingress, error) {

	const pageSize int64 = 100

	ingresses := make([]netv1.Ingress, 0)
	var continueToken string

	for {
		opts := metav1.ListOptions{
			Limit:    pageSize,
			Continue: continueToken,
		}

		list, err := cfg.clientSet.NetworkingV1().Ingresses(cfg.NameSpace).List(context.TODO(), opts)
		if err != nil {
			return nil, err
		}

		ingresses = append(ingresses, list.Items...)

		if list.Continue == "" {
			break
		}

		continueToken = list.Continue
	}

	return ingresses, nil
}
