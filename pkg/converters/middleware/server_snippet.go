package middleware

import (
	"strings"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
)

/* ---------------- SERVER SNIPPET ---------------- */

// ServerSnippet handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/server-snippet"
func ServerSnippet(ctx configs.Context) {
	ctx.Log.Debug("running converter ServerSnippet")

	snippet, ok := ctx.Annotations[string(models.ServerSnippet)]
	if !ok || strings.TrimSpace(snippet) == "" {
		return
	}

	// 1) Header-only server-snippet (heuristic)
	if isOnlyAddHeader(snippet) {
		warningMessage := "server-snippet contains only add_header directives. " +
			"These were not auto-converted because server-snippet applies " +
			"at NGINX server scope. Consider moving them to " +
			"nginx.ingress.kubernetes.io/configuration-snippet " +
			"or converting them manually to a Traefik Headers middleware."

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)

		ctx.ReportSkipped(string(models.ServerSnippet), warningMessage)

		return
	}

	// 2) Header buffer tuning (static Traefik config)
	if strings.Contains(snippet, "client_header_buffer_size") ||
		strings.Contains(snippet, "large_client_header_buffers") {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"server-snippet configures request header buffer sizes. "+
				"Traefik does not support per-route header buffer tuning. "+
				"Equivalent settings must be configured globally on entryPoints "+
				"(e.g. http.maxHeaderBytes) in Traefik static configuration.",
		)

		return
	}

	// 3) Timeout tuning (proxy / send timeouts)
	if strings.Contains(snippet, "proxy_read_timeout") ||
		strings.Contains(snippet, "proxy_send_timeout") ||
		strings.Contains(snippet, "send_timeout") {
		warningMessage := "server-snippet configures timeout settings (proxy/send timeouts). " +
			"These cannot be set per-route in Traefik. Consider using ServersTransport " +
			"(e.g. forwardingTimeouts) in dynamic configuration or static config instead."

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)
		ctx.ReportSkipped(string(models.ServerSnippet), warningMessage)

		return
	}

	// 4) TLS knobs (ssl_* / proxy_ssl_*)
	if strings.Contains(snippet, "ssl_") || strings.Contains(snippet, "proxy_ssl_") {
		warningMessage := "server-snippet configures TLS-related directives. " +
			"These cannot be safely auto-converted. In Traefik, use TLSOption " +
			"and/or ServersTransport for TLS configuration."

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)
		ctx.ReportSkipped(string(models.ServerSnippet), warningMessage)

		return
	}

	// 5) Rate limiting (limit_req / limit_conn)
	if strings.Contains(snippet, "limit_req") || strings.Contains(snippet, "limit_conn") {
		warningMessage := "server-snippet configures NGINX rate limiting (limit_req/limit_conn). " +
			"Traefik provides a RateLimit middleware, but semantics differ and this cannot be " +
			"auto-converted safely."

		ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)
		ctx.ReportSkipped(string(models.ServerSnippet), warningMessage)

		return
	}

	warningMessage := "server-snippet injects raw NGINX server configuration which has no Traefik equivalent; skipped"

	ctx.Result.Warnings = append(ctx.Result.Warnings, warningMessage)

	ctx.ReportSkipped(string(models.ServerSnippet), warningMessage)
}

func isOnlyAddHeader(snippet string) bool {
	lines := strings.Split(snippet, "\n")
	for _, l := range lines {
		line := strings.TrimSpace(l)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "add_header") {
			return false
		}
	}

	return true
}
