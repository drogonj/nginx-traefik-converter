package configs

import "sigs.k8s.io/controller-runtime/pkg/client"

// Result holds the translated configs for a nginx ingress.
type Result struct {
	Middlewares   []client.Object
	IngressRoutes []client.Object
	TLSOptions    []client.Object
	TLSOptionRefs map[string]string // ingressName → tlsOptionName
	Warnings      []string
}

// NewResult returns new instance of Result
func NewResult() *Result {
	return &Result{}
}
