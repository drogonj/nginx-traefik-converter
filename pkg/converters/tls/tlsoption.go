// pkg/convert/tlsoption.go
package tls

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	traefik "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func EmitTLSOption(ctx configs.Context, secretName string) {
	if ctx.Result.TLSOptionRefs == nil {
		ctx.Result.TLSOptionRefs = make(map[string]string)
	}

	name := ctx.IngressName + "-mtls"

	tlsOpt := &traefik.TLSOption{
		TypeMeta: metav1.TypeMeta{
			APIVersion: traefik.SchemeGroupVersion.String(),
			Kind:       "TLSOption",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ctx.Namespace,
		},
		Spec: traefik.TLSOptionSpec{
			ClientAuth: traefik.ClientAuth{
				ClientAuthType: "RequireAndVerifyClientCert",
				SecretNames:    []string{secretName},
			},
		},
	}

	ctx.Result.TLSOptions = append(ctx.Result.TLSOptions, tlsOpt)
	ctx.Result.TLSOptionRefs[ctx.IngressName] = name

	ctx.Result.Warnings = append(ctx.Result.Warnings,
		"auth-tls-secret must contain CA certificates only; server cert secrets cannot be reused",
		"CA certificates must be mounted into Traefik via static configuration",
	)
}

func ApplyTLSOption(ir *traefik.IngressRoute, ctx configs.Context) {
	if opt, ok := ctx.Result.TLSOptionRefs[ctx.IngressName]; ok {
		ir.Spec.TLS = &traefik.TLS{
			Options: &traefik.TLSOptionRef{
				Name: opt,
			},
		}
	}
}
