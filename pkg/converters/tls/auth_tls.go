package tls

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
)

func HandleAuthTLSVerifyClient(ctx configs.Context) {
	verify := ctx.Annotations["nginx.ingress.kubernetes.io/auth-tls-verify-client"]
	if verify != "on" && verify != "true" {
		return
	}

	secret := ctx.Annotations["nginx.ingress.kubernetes.io/auth-tls-secret"]
	if secret == "" {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"auth-tls-verify-client is enabled but auth-tls-secret is missing",
		)
		return
	}

	EmitTLSOption(ctx, secret)
}
