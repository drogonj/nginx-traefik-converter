package configs

import (
	"fmt"
	"strings"
)

var helmLabelIndicators = []struct{ key, value string }{
	{"heritage", "Helm"},
	{"app.kubernetes.io/managed-by", "Helm"},
	{"helm.sh/", ""},
}

// helmAnnotationPrefixes lists annotation prefixes that indicate Helm management.
var helmAnnotationPrefixes = []string{
	"meta.helm.sh/",
}

// DetectHelmIndicators checks labels and annotations for Helm management
func DetectHelmIndicators(labels, annotations map[string]string) []string {
	var found []string

	for _, ind := range helmLabelIndicators {
		if strings.HasSuffix(ind.key, "/") {
			// Prefix match — any label whose key starts with this prefix qualifies.
			for k := range labels {
				if strings.HasPrefix(k, ind.key) {
					found = append(found, fmt.Sprintf("label %q=%q", k, labels[k]))
				}
			}
		} else if v, ok := labels[ind.key]; ok {
			if ind.value == "" || strings.EqualFold(v, ind.value) {
				found = append(found, fmt.Sprintf("label %q=%q", ind.key, v))
			}
		}
	}

	for _, prefix := range helmAnnotationPrefixes {
		for k, v := range annotations {
			if strings.HasPrefix(k, prefix) {
				found = append(found, fmt.Sprintf("annotation %q=%q", k, v))
			}
		}
	}

	return found
}
