package convert

import (
	"fmt"
	"strings"
)

//func detectBackendProtocol(ctx Context) (string, bool, error) {
//	proto := ctx.Annotations["nginx.ingress.kubernetes.io/backend-protocol"]
//
//	if ctx.Annotations["nginx.ingress.kubernetes.io/grpc-backend"] == "true" && proto == "" {
//		proto = "GRPC"
//	}
//
//	switch strings.ToUpper(proto) {
//	case "", "HTTP":
//		return "http", false, nil
//	case "HTTPS":
//		return "https", true, nil
//	case "GRPC":
//		return "h2c", true, nil
//	case "GRPCS":
//		return "https", true, nil
//	default:
//		return "", false, fmt.Errorf("unsupported backend-protocol: %s", proto)
//	}
//}

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

func needsIngressRoute(ann map[string]string) bool {
	if ann["nginx.ingress.kubernetes.io/grpc-backend"] == "true" {
		return true
	}

	if _, ok := ann["nginx.ingress.kubernetes.io/backend-protocol"]; ok {
		return true
	}

	return false
}
