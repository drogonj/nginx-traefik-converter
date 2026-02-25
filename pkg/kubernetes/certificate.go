package kubernetes

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// certManagerGVR is the GroupVersionResource for cert-manager Certificate CRDs.
var certManagerGVR = schema.GroupVersionResource{
	Group:    "cert-manager.io",
	Version:  "v1",
	Resource: "certificates",
}

// FindCertificateBySecret returns the first Certificate in the given namespace
// whose spec.secretName matches the provided value.
// Returns (nil, nil) when no matching Certificate is found.
//
// Results are cached per namespace so that multiple TLS entries in the same
// namespace only trigger a single LIST call against the API server.
//
// This approach (LIST + filter) is used instead of GET-by-name because
// cert-manager does not guarantee that the Certificate name matches the
// secretName — users may have created Certificates manually with arbitrary
// names.
func (cfg *Config) FindCertificateBySecret(namespace, secretName string) (*unstructured.Unstructured, error) {
	if cfg.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client is not initialised; cannot look up Certificate resources")
	}

	items, err := cfg.listCertificatesCached(namespace)
	if err != nil {
		return nil, err
	}

	// items is nil when the CRD is absent — caller will fall back to generation.
	if items == nil {
		return nil, nil
	}

	for i := range items {
		cert := &items[i]

		specSecret, found, _ := unstructured.NestedString(cert.Object, "spec", "secretName")
		if found && specSecret == secretName {
			return cert, nil
		}
	}

	return nil, nil
}

// listCertificatesCached returns all Certificate items in the namespace,
// caching the result so repeated calls for the same namespace reuse the
// first LIST response.
//
// Returns (nil, nil) when the cert-manager CRD is not installed.
func (cfg *Config) listCertificatesCached(namespace string) ([]unstructured.Unstructured, error) {
	if cfg.certCache == nil {
		cfg.certCache = make(map[string]*certCacheEntry)
	}

	if entry, ok := cfg.certCache[namespace]; ok {
		return entry.items, entry.err
	}

	items, err := cfg.doListCertificates(namespace)
	cfg.certCache[namespace] = &certCacheEntry{items: items, err: err}

	return items, err
}

// doListCertificates performs the actual LIST call to the API server.
func (cfg *Config) doListCertificates(namespace string) ([]unstructured.Unstructured, error) {
	list, err := cfg.dynamicClient.Resource(certManagerGVR).Namespace(namespace).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		// The cert-manager CRD may not be installed in the cluster.
		// Treat this gracefully: return nil so the caller can fall back
		// to generating a Certificate from annotations.
		if apierrors.IsNotFound(err) || apimeta.IsNoMatchError(err) {
			if cfg.logger != nil {
				cfg.logger.Warn("cert-manager Certificate CRD not found in cluster; skipping extraction")
			}

			return nil, nil
		}

		return nil, fmt.Errorf("listing cert-manager Certificates in namespace %q: %w", namespace, err)
	}

	return list.Items, nil
}

// SanitizeCertificateForGitOps removes cluster-specific and transient fields
// from a live Certificate resource so it can be safely committed to Git and
// managed declaratively by ArgoCD (or similar).
func SanitizeCertificateForGitOps(cert *unstructured.Unstructured) {
	// Remove server-set metadata that causes drift detection or conflicts.
	unstructured.RemoveNestedField(cert.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(cert.Object, "metadata", "uid")
	unstructured.RemoveNestedField(cert.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(cert.Object, "metadata", "generation")
	unstructured.RemoveNestedField(cert.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(cert.Object, "metadata", "selfLink")

	// Critical: Remove ownerReferences. The Certificate extracted from the
	// cluster points to the Ingress that will be deleted during migration.
	// Without this cleanup cert-manager's garbage collector would delete the
	// Certificate (and its Secret) as soon as the Ingress is removed.
	unstructured.RemoveNestedField(cert.Object, "metadata", "ownerReferences")

	// Remove verbose last-applied-configuration annotation injected by kubectl.
	annotations, _, _ := unstructured.NestedMap(cert.Object, "metadata", "annotations")
	if annotations != nil {
		delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")

		if len(annotations) == 0 {
			unstructured.RemoveNestedField(cert.Object, "metadata", "annotations")
		} else {
			_ = unstructured.SetNestedMap(cert.Object, annotations, "metadata", "annotations")
		}
	}

	// Remove labels that become misleading once the Certificate is extracted
	// from its original lifecycle (Helm release, operator, etc.) and managed
	// declaratively in Git.
	cleanStaleLabels(cert)

	// Remove runtime status — it is not declarative.
	unstructured.RemoveNestedField(cert.Object, "status")
}

// staleLabelsToRemove lists label keys (or prefixes ending with "/") that
// should be stripped from extracted resources.  They either reference the
// previous lifecycle manager or contain values that will drift over time.
var staleLabelsToRemove = []string{
	"helm.sh/",                     // helm.sh/chart, helm.sh/heritage, …
	"app.kubernetes.io/managed-by", // "Helm" → no longer true
	"app.kubernetes.io/version",    // version of the app — goes stale
}

// cleanStaleLabels removes Helm-specific and version labels that are no longer
// accurate once the resource is detached from its original release.
func cleanStaleLabels(obj *unstructured.Unstructured) {
	labels := obj.GetLabels()
	if len(labels) == 0 {
		return
	}

	filtered := make(map[string]string, len(labels))

	for k, v := range labels {
		if isStaleLabel(k) {
			continue
		}

		filtered[k] = v
	}

	if len(filtered) == 0 {
		unstructured.RemoveNestedField(obj.Object, "metadata", "labels")
	} else {
		obj.SetLabels(filtered)
	}
}

// isStaleLabel returns true when the key matches one of the entries in
// staleLabelsToRemove.  Entries ending with "/" are treated as prefixes.
func isStaleLabel(key string) bool {
	for _, pattern := range staleLabelsToRemove {
		if pattern[len(pattern)-1] == '/' {
			// prefix match
			if len(key) >= len(pattern) && key[:len(pattern)] == pattern {
				return true
			}
		} else if key == pattern {
			return true
		}
	}

	return false
}
