package convert

import "sigs.k8s.io/controller-runtime/pkg/client"

type Result struct {
	Middlewares   []client.Object
	IngressRoutes []client.Object
	TLSOptions    []client.Object
	TLSOptionRefs map[string]string // ingressName → tlsOptionName
	Warnings      []string
}
