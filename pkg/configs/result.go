package configs

import "sigs.k8s.io/controller-runtime/pkg/client"

// Result holds the translated configs for a nginx ingress.
type Result struct {
	Middlewares   []client.Object   `yaml:"middlewares,omitempty"     json:"middlewares,omitempty"`
	IngressRoutes []client.Object   `yaml:"ingress_routes,omitempty"  json:"ingress_routes,omitempty"`
	TLSOptions    []client.Object   `yaml:"tls_options,omitempty"     json:"tls_options,omitempty"`
	TLSOptionRefs map[string]string `yaml:"tls_option_refs,omitempty" json:"tls_option_refs,omitempty"`
	Warnings      []string          `yaml:"warnings,omitempty"        json:"warnings,omitempty"`
	IngressReport IngressReport     `yaml:"ingress_report,omitempty"  json:"ingress_report,omitempty"`
	// Report        GlobalReport      `yaml:"report,omitempty"         json:"report,omitempty"`
}

// NewResult returns new instance of Result.
func NewResult() *Result {
	return &Result{}
}
