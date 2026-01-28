package convert

func Run(ctx Context) error {
	rewriteTarget(ctx)
	sslRedirect(ctx)
	basicAuth(ctx)

	if err := cors(ctx); err != nil {
		return err
	}
	rateLimit(ctx)

	if err := bodySize(ctx); err != nil {
		return err
	}

	extraAnnotations(ctx)
	handleAuthTLSVerifyClient(ctx)
	configurationSnippet(ctx)

	if needsIngressRoute(ctx.Annotations) {
		if err := buildIngressRoute(ctx); err != nil {
			ctx.Result.Warnings = append(ctx.Result.Warnings, err.Error())
		}
	}

	warnings(ctx)

	return nil
}
