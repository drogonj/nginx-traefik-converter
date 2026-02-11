package middleware

import (
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
)

/* ---------------- UNSUPPORTED/REDUNDANT ANNOTATIONS ---------------- */

// ExtraAnnotations handles the below unsupported annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/proxy-buffering"
//   - "nginx.ingress.kubernetes.io/service-upstream"
//   - "nginx.ingress.kubernetes.io/enable-opentracing"
//   - "nginx.ingress.kubernetes.io/enable-opentelemetry"
//   - "nginx.ingress.kubernetes.io/backend-protocol"
//   - "nginx.ingress.kubernetes.io/grpc-backend"
func ExtraAnnotations(ctx configs.Context) {
	ctx.Log.Debug("running converter ExtraAnnotations")

	if _, ok := ctx.Annotations[string(models.ClientHeaderBufferSize)]; ok {
		msg := `'client-header-buffer-size' is not supported per-Ingress in Traefik;
configure it globally using:
entryPoints.<name>.http.maxHeaderBytes in Traefik static configuration`

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportIgnored(string(models.ClientHeaderBufferSize), msg)
	}

	if _, ok := ctx.Annotations[string(models.LargeClientHeaderBuffers)]; ok {
		msg := `'large-client-header-buffers' is not supported per-Ingress in Traefik;
Traefik only supports a global limit
via entryPoints.<name>.http.maxHeaderBytes in static configuration`

		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportIgnored(string(models.LargeClientHeaderBuffers), msg)
	}

	if ctx.Annotations[string(models.ServiceUpstream)] == "true" {
		warningMessage := "service-upstream=true is default behavior in Traefik"

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)
		ctx.ReportIgnored(string(models.ServiceUpstream), warningMessage)
	}

	if ctx.Annotations[string(models.EnableOpentracing)] == "true" {
		warningMessage := "enable-opentracing is global in Traefik and cannot be enabled per Ingress"

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)
		ctx.ReportWarning(string(models.EnableOpentracing), warningMessage)
	}

	if ctx.Annotations[string(models.EnableOpentelemetry)] == "true" {
		warningMessage := "enable-opentelemetry must be configured globally in Traefik static config"

		ctx.Result.Warnings = append(
			ctx.Result.Warnings,
			warningMessage+`tracing:
  otlp:
    grpc:
      endpoint: otel-collector:4317`,
		)

		ctx.ReportWarning(string(models.EnableOpentelemetry), warningMessage)
	}

	if v := ctx.Annotations[string(models.BackendProtocol)]; v != "" {
		warningMessage := "backend-protocol must be applied to IngressRoute service scheme, check for generated ingressroutes.yaml"

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)
		ctx.ReportWarning(string(models.BackendProtocol), warningMessage)
	}

	if ctx.Annotations[string(models.GrpcBackend)] == "true" {
		warningMessage := "grpc-backend requires IngressRoute service scheme h2c or https+h2, check for generated ingressroutes.yaml"

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)
		ctx.ReportWarning(string(models.GrpcBackend), warningMessage)
	}
}
