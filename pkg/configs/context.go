package configs

import (
	netv1 "k8s.io/api/networking/v1"
	"log/slog"
)

type Context struct {
	Ingress     *netv1.Ingress    `yaml:"ingress,omitempty" json:"ingress,omitempty"`
	IngressName string            `yaml:"ingress_name,omitempty" json:"ingress_name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Result      *Result           `yaml:"result,omitempty" json:"result,omitempty"`
	log         *slog.Logger
}

func New(ingress *netv1.Ingress, result *Result) *Context {
	return &Context{
		Ingress:     ingress,
		IngressName: ingress.Name,
		Namespace:   ingress.Namespace,
		Annotations: ingress.Annotations,
		Result:      result,
	}
}
