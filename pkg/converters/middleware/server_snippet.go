package middleware

import (
	"strings"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/converters/models"
)

/* ---------------- PROXY REDIRECT ---------------- */

// ServerSnippet handles the below annotations.
// Annotations:
//   - "nginx.ingress.kubernetes.io/server-snippet"
func ServerSnippet(ctx configs.Context) {
	snippet, ok := ctx.Annotations[string(models.ServerSnippet)]
	if !ok || strings.TrimSpace(snippet) == "" {
		return
	}

	// Header-only server-snippet (heuristic)
	if isOnlyAddHeader(snippet) {
		ctx.Result.Warnings = append(ctx.Result.Warnings,
			"server-snippet contains only add_header directives. "+
				"These were not auto-converted because server-snippet applies "+
				"at NGINX server scope. Consider moving them to "+
				"nginx.ingress.kubernetes.io/configuration-snippet "+
				"or converting them manually to a Traefik Headers middleware.",
		)

		return
	}

	// Detect header buffer tuning
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

	ctx.Result.Warnings = append(ctx.Result.Warnings,
		"server-snippet injects raw NGINX server configuration which has no Traefik equivalent; skipped",
	)

	ctx.ReportSkipped(ctx.Annotations[string(models.ServerSnippet)], "injects raw NGINX server-level configuration; no Traefik equivalent")
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
