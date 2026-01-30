package ingressroute

import (
	"strings"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/errors"
)

func resolveScheme(
	annotations map[string]string,
) (string, error) {
	if annotations["nginx.ingress.kubernetes.io/grpc-backend"] == "true" {
		return "h2c", nil
	}

	switch strings.ToUpper(
		annotations["nginx.ingress.kubernetes.io/backend-protocol"],
	) {
	case "", "HTTP":
		return "http", nil
	case "HTTPS":
		return "https", nil
	case "GRPC":
		return "h2c", nil
	case "GRPCS":
		return "https", nil
	default:
		return "", &errors.ConverterError{Message: "unsupported backend-protocol"}
	}
}

func entryPointsForScheme(scheme string) []string {
	switch scheme {
	case "https":
		return []string{"websecure"}
	case "h2c":
		return []string{"web"}
	default:
		return []string{"web"}
	}
}

// NeedsIngressRoute makes the decision on requirement of ingress routes.
func NeedsIngressRoute(ann map[string]string) bool {
	if ann["nginx.ingress.kubernetes.io/grpc-backend"] == "true" {
		return true
	}

	if _, ok := ann["nginx.ingress.kubernetes.io/backend-protocol"]; ok {
		return true
	}

	return false
}
