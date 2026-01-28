package convert

func handleAuthTLSVerifyClient(ctx Context) {
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

	emitTLSOption(ctx, secret)
}
