package configs

import (
	"log/slog"

	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CertificateLookup provides cluster-level access to cert-manager Certificate
// resources. Implementations are expected to query the Kubernetes API using
// the dynamic client. A nil CertificateLookup means cluster lookups are
// unavailable (e.g. offline / file-based mode) and the converter should fall
// back to generating Certificates from annotations only.
type CertificateLookup interface {
	// FindCertificateBySecret lists Certificate resources in the given
	// namespace and returns the first one whose spec.secretName matches.
	// Returns nil (no error) when no matching Certificate is found.
	FindCertificateBySecret(namespace, secretName string) (*unstructured.Unstructured, error)
}

// Context holds the necessary info required to run the converters.
type Context struct {
	Ingress     *netv1.Ingress    `yaml:"ingress,omitempty" json:"ingress,omitempty"`
	IngressName string            `yaml:"ingress_name,omitempty" json:"ingress_name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Result      *Result           `yaml:"result,omitempty" json:"result,omitempty"`
	Options     *Options          `yaml:"options,omitempty" json:"options,omitempty"`
	CertLookup  CertificateLookup `yaml:"-" json:"-"`
	Log         *slog.Logger
}

// New returns a new instance of Context when invoked.
func New(ingress *netv1.Ingress, result *Result, options *Options, logger *slog.Logger) *Context {
	return &Context{
		Ingress:     ingress,
		IngressName: ingress.Name,
		Namespace:   ingress.Namespace,
		Annotations: ingress.Annotations,
		Result:      result,
		Options:     options,
		Log:         logger,
	}
}
