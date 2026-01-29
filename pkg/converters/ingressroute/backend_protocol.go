package ingressroute

import (
	"fmt"
	"strings"
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
		return "", fmt.Errorf("unsupported backend-protocol")
	}
}

func ingressRouteName(base, scheme string) string {
	switch scheme {
	case "h2c":
		return base + "-grpc"
	case "https":
		return base + "-https"
	default:
		return base
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

func NeedsIngressRoute(ann map[string]string) bool {
	if ann["nginx.ingress.kubernetes.io/grpc-backend"] == "true" {
		return true
	}

	if _, ok := ann["nginx.ingress.kubernetes.io/backend-protocol"]; ok {
		return true
	}

	return false
}
