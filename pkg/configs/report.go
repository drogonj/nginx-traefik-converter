package configs

// AnnotationStatus represents the migration outcome of a single
// NGINX Ingress annotation during conversion to Traefik.
//
// It is used in reports to indicate whether an annotation was:
//   - fully converted,
//   - converted with warnings,
//   - skipped because it cannot be safely migrated, or
//   - ignored because it is not relevant for Traefik.

type AnnotationStatus string

const (
	// AnnotationConverted indicates that the annotation was successfully
	// converted into equivalent Traefik configuration.
	AnnotationConverted AnnotationStatus = "converted"

	// AnnotationSkipped indicates that the annotation was detected but could
	// not be safely converted and therefore requires manual migration.
	AnnotationSkipped AnnotationStatus = "skipped"

	// AnnotationWarned indicates that the annotation was converted, but the
	// resulting Traefik configuration may differ in behavior and should be
	// reviewed by the user.
	AnnotationWarned AnnotationStatus = "warning"

	// AnnotationIgnored indicates that the annotation was detected but was
	// intentionally ignored because it is not applicable or has no effect
	// in Traefik.
	AnnotationIgnored AnnotationStatus = "ignored"
)

// AnnotationReportEntry represents the migration result of a single
// NGINX Ingress annotation.
type AnnotationReportEntry struct {
	// Name is the full annotation key, for example:
	// "nginx.ingress.kubernetes.io/rewrite-target".
	Name string `yaml:"name,omitempty"    json:"name,omitempty"`

	// Status indicates the outcome of the conversion for this annotation.
	Status AnnotationStatus `yaml:"status,omitempty"  json:"status,omitempty"`

	// Message contains an optional human-readable explanation, typically
	// used for warnings and skipped annotations.
	Message string `yaml:"message,omitempty" json:"message,omitempty"`
}

// IngressReport contains the migration report for a single Kubernetes Ingress.
type IngressReport struct {
	// Namespace is the namespace of the Ingress resource.
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`

	// Name is the name of the Ingress resource.
	Name string `yaml:"name,omitempty"      json:"name,omitempty"`

	// Entries is the list of per-annotation migration results.
	Entries []AnnotationReportEntry `yaml:"entries,omitempty"   json:"entries,omitempty"`
}

// GlobalReport aggregates migration reports for all processed Ingresses.
type GlobalReport struct {
	// Ingresses is the list of per-Ingress migration reports.
	Ingresses []IngressReport `yaml:"ingresses,omitempty" json:"ingresses,omitempty"`
}

// StartIngressReport initializes a new per-Ingress report in the current
// conversion context. It should be called once at the beginning of processing
// each Ingress.
func (ctx *Context) StartIngressReport(ns, name string) {
	ctx.Result.IngressReport = IngressReport{
		Namespace: ns,
		Name:      name,
		Entries:   nil,
	}
}

// addReport appends a new annotation report entry to the current Ingress report.
// It is an internal helper used by the public Report* methods.
func (ctx *Context) addReport(name string, status AnnotationStatus, msg string) {
	ctx.Result.IngressReport.Entries = append(
		ctx.Result.IngressReport.Entries,
		AnnotationReportEntry{
			Name:    name,
			Status:  status,
			Message: msg,
		},
	)
}

// ReportConverted records that the given annotation was successfully converted
// into Traefik configuration without requiring manual action.
func (ctx *Context) ReportConverted(name string) {
	ctx.addReport(name, AnnotationConverted, "")
}

// ReportSkipped records that the given annotation was detected but could not be
// safely converted and therefore requires manual migration.
func (ctx *Context) ReportSkipped(name, msg string) {
	ctx.addReport(name, AnnotationSkipped, msg)
}

// ReportWarning records that the given annotation was converted, but the result
// may differ in behavior from the original NGINX configuration and should be
// reviewed by the user.
func (ctx *Context) ReportWarning(name, msg string) {
	ctx.addReport(name, AnnotationWarned, msg)
}

// ReportIgnored records that the given annotation was detected but intentionally
// ignored because it is not relevant or has no effect in Traefik.
func (ctx *Context) ReportIgnored(name string, msg string) {
	ctx.addReport(name, AnnotationIgnored, msg)
}
