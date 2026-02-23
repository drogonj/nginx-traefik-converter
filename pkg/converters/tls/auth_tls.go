package tls

import (
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
)

// HandleAuthTLSVerifyClient is responsible for handling TLS configs of nginx annotations, handles the below.
// Annotations:
//   - "nginx.ingress.kubernetes.io/auth-tls-verify-client"
//   - "nginx.ingress.kubernetes.io/auth-tls-secret"
func HandleAuthTLSVerifyClient(ctx configs.Context) {
	ctx.Log.Debug("running converter HandleAuthTLSVerifyClient")

	verify := ctx.Annotations[string(models.AuthTLSVerifyClient)]
	if verify == "" || verify == "off" || verify == "false" {
		return
	}

	secret := ctx.Annotations[string(models.AuthTLSSecret)]
	if secret == "" {
		msg := "auth-tls-verify-client is enabled but auth-tls-secret is missing"

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(string(models.AuthTLSVerifyClient), msg)
		ctx.ReportSkipped(string(models.AuthTLSSecret), msg)

		return
	}

	var clientAuthType string

	switch verify {
	case "on", "true":
		clientAuthType = "RequireAndVerifyClientCert"
	case "optional":
		clientAuthType = "VerifyClientCertIfGiven"
	case "optional_no_ca":
		msg := "auth-tls-verify-client=optional_no_ca has no safe Traefik equivalent; skipped"
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(string(models.AuthTLSVerifyClient), msg)

		return
	default:
		msg := "unsupported value for auth-tls-verify-client: " + verify
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped(string(models.AuthTLSVerifyClient), msg)

		return
	}

	emitTLSOption(ctx, secret, clientAuthType)
}
