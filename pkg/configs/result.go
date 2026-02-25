package configs

import (
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Result holds the translated configs for a nginx ingress.
type Result struct {
	Middlewares   []*traefik.Middleware   `yaml:"middlewares,omitempty"     json:"middlewares,omitempty"`
	IngressRoutes []*traefik.IngressRoute `yaml:"ingress_routes,omitempty"  json:"ingress_routes,omitempty"`
	TLSOptions    []*traefik.TLSOption    `yaml:"tls_options,omitempty"     json:"tls_options,omitempty"`
	TLSOptionRefs map[string]string       `yaml:"tls_option_refs,omitempty" json:"tls_option_refs,omitempty"`

	// Certificates holds cert-manager Certificate resources extracted from
	// the cluster or generated from Ingress annotations. Stored as
	// Unstructured to avoid pulling in the cert-manager Go module.
	Certificates []*unstructured.Unstructured `yaml:"certificates,omitempty"    json:"certificates,omitempty"`

	Warnings      []string      `yaml:"warnings,omitempty"        json:"warnings,omitempty"`
	IngressReport IngressReport `yaml:"ingress_report,omitempty"  json:"ingress_report,omitempty"`
	// Report        GlobalReport      `yaml:"report,omitempty"         json:"report,omitempty"`
}

// NewResult returns new instance of Result.
func NewResult() *Result {
	return &Result{}
}
