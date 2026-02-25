// Package certificate handles the extraction or generation of cert-manager
// Certificate resources during nginx-to-traefik migration.
//
// Strategy (extract-first, generate-fallback):
//  1. For each spec.tls[] entry with a secretName, attempt to extract the
//     live Certificate from the cluster via the CertificateLookup interface.
//  2. If found, sanitize it for GitOps (remove ownerReferences, status, etc.).
//  3. If not found, fall back to generating a Certificate from cert-manager
//     annotations on the Ingress.
//  4. If neither a live Certificate nor annotations exist, emit a warning.
//
// Deduplication is handled by secretName to avoid producing the same
// Certificate twice when multiple Ingresses share a TLS secret.
package certificate

import (
	"fmt"

	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/configs"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/converters/models"
	"github.com/nikhilsbhat/nginx-traefik-converter/pkg/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	certManagerAPIVersion = "cert-manager.io/v1"
	certManagerKind       = "Certificate"

	defaultIssuerGroup = "cert-manager.io"
)

// ExtractOrGenerate processes each spec.tls[] entry of the Ingress and
// produces a cert-manager Certificate resource — either extracted from the
// cluster or generated from annotations.
func ExtractOrGenerate(ctx configs.Context) {
	ing := ctx.Ingress
	if len(ing.Spec.TLS) == 0 {
		return
	}

	seen := make(map[string]struct{})

	for _, tlsEntry := range ing.Spec.TLS {
		secretName := tlsEntry.SecretName
		if secretName == "" {
			ctx.Result.Warnings = append(ctx.Result.Warnings,
				fmt.Sprintf("TLS entry with hosts %v has no secretName; skipping Certificate extraction", tlsEntry.Hosts))

			continue
		}

		// Dedup: skip if we already processed this secretName.
		if _, exists := seen[secretName]; exists {
			continue
		}

		seen[secretName] = struct{}{}

		// Step 1: try to extract the live Certificate from the cluster.
		if extracted := tryExtract(ctx, secretName); extracted {
			continue
		}

		// Step 2: fall back to generating from annotations.
		if generated := tryGenerate(ctx, secretName, tlsEntry.Hosts); generated {
			continue
		}

		// Step 3: neither extraction nor generation succeeded.
		msg := fmt.Sprintf("no Certificate found in cluster for secret %q and no cert-manager issuer annotation present; manual Certificate creation required", secretName)
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportSkipped("cert-manager/certificate", msg)
	}
}

// tryExtract attempts to retrieve the Certificate from the cluster using the
// CertificateLookup interface. Returns true if a Certificate was found and
// added to the result.
func tryExtract(ctx configs.Context, secretName string) bool {
	if ctx.CertLookup == nil {
		return false
	}

	cert, err := ctx.CertLookup.FindCertificateBySecret(ctx.Namespace, secretName)
	if err != nil {
		msg := fmt.Sprintf("failed to look up Certificate for secret %q: %v", secretName, err)
		ctx.Result.Warnings = append(ctx.Result.Warnings, msg)
		ctx.ReportWarning("cert-manager/certificate", msg)

		return false
	}

	if cert == nil {
		return false
	}

	// Sanitize the live resource for GitOps.
	kubernetes.SanitizeCertificateForGitOps(cert)

	ctx.Result.Certificates = append(ctx.Result.Certificates, cert)
	ctx.ReportConverted("cert-manager/certificate")

	ctx.Log.Info("extracted Certificate from cluster",
		"secret", secretName,
		"certificate", cert.GetName(),
	)

	return true
}

// tryGenerate builds a Certificate from cert-manager annotations on the
// Ingress. Returns true if a Certificate was generated and added to the result.
func tryGenerate(ctx configs.Context, secretName string, hosts []string) bool {
	// Resolve issuer: cluster-issuer takes precedence over namespaced issuer.
	issuerName, issuerKind := resolveIssuer(ctx.Annotations)
	if issuerName == "" {
		return false
	}

	issuerGroup := ctx.Annotations[string(models.CertManagerIssuerGroup)]
	if issuerGroup == "" {
		issuerGroup = defaultIssuerGroup
	}

	// Build the Certificate as Unstructured.
	cert := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": certManagerAPIVersion,
			"kind":       certManagerKind,
			"metadata": map[string]interface{}{
				"name":      secretName,
				"namespace": ctx.Namespace,
			},
			"spec": map[string]interface{}{
				"secretName": secretName,
				"dnsNames":   toInterfaceSlice(hosts),
				"issuerRef": map[string]interface{}{
					"name":  issuerName,
					"kind":  issuerKind,
					"group": issuerGroup,
				},
			},
		},
	}

	// Apply optional fields from annotations.
	applyOptionalFields(cert, ctx.Annotations)

	ctx.Result.Certificates = append(ctx.Result.Certificates, cert)

	ctx.ReportConverted(string(models.CertManagerClusterIssuer))

	ctx.Log.Info("generated Certificate from annotations",
		"secret", secretName,
		"issuer", issuerName,
		"kind", issuerKind,
	)

	return true
}

// resolveIssuer determines the issuer name and kind from annotations.
// cluster-issuer takes precedence over namespaced issuer.
func resolveIssuer(annotations map[string]string) (name, kind string) {
	if v := annotations[string(models.CertManagerClusterIssuer)]; v != "" {
		return v, "ClusterIssuer"
	}

	if v := annotations[string(models.CertManagerIssuer)]; v != "" {
		kind = annotations[string(models.CertManagerIssuerKind)]
		if kind == "" {
			kind = "Issuer"
		}

		return v, kind
	}

	return "", ""
}

// applyOptionalFields sets optional Certificate spec fields when the
// corresponding cert-manager annotations are present.
func applyOptionalFields(cert *unstructured.Unstructured, annotations map[string]string) {
	if v := annotations[string(models.CertManagerCommonName)]; v != "" {
		_ = unstructured.SetNestedField(cert.Object, v, "spec", "commonName")
	}

	if v := annotations[string(models.CertManagerDuration)]; v != "" {
		_ = unstructured.SetNestedField(cert.Object, v, "spec", "duration")
	}

	if v := annotations[string(models.CertManagerRenewBefore)]; v != "" {
		_ = unstructured.SetNestedField(cert.Object, v, "spec", "renewBefore")
	}
}

// toInterfaceSlice converts a []string to []interface{} for use with
// unstructured.SetNestedSlice.
func toInterfaceSlice(ss []string) []interface{} {
	out := make([]interface{}, len(ss))
	for i, s := range ss {
		out[i] = s
	}

	return out
}
