// Package render provides human-friendly output renderers for migration reports.
// It can print per-Ingress and global summaries in either table format (using
// tablewriter) or plain text format, depending on configuration.
package render

import (
	"fmt"
	"os"
	"strconv"

	"github.com/nikhilsbhat/ingress-traefik-converter/pkg/configs"
	"github.com/olekukonko/tablewriter"
)

// SummaryCounts represents aggregated counts of annotation conversion outcomes.
// It is used for both per-Ingress summaries and global summaries.
type SummaryCounts struct {
	// Converted is the number of annotations successfully converted.
	Converted int `yaml:"converted,omitempty" json:"converted,omitempty"`

	// Warnings is the number of annotations converted with warnings.
	Warnings int `yaml:"warnings,omitempty"  json:"warnings,omitempty"`

	// Skipped is the number of annotations that could not be converted and
	// require manual migration.
	Skipped int `yaml:"skipped,omitempty"   json:"skipped,omitempty"`

	// Ignored is the number of annotations that were intentionally ignored.
	Ignored int `yaml:"ignored,omitempty"   json:"ignored,omitempty"`
}

// Config controls how reports are rendered.
type Config struct {
	// Table determines whether output should be rendered as tables (true)
	// or as plain text (false).
	Table bool `yaml:"table,omitempty" json:"table,omitempty"`
}

// statusLabel maps annotation statuses to human-readable labels for display.
var statusLabel = map[configs.AnnotationStatus]string{
	configs.AnnotationConverted: "Converted",
	configs.AnnotationWarned:    "Warning",
	configs.AnnotationSkipped:   "Skipped",
	configs.AnnotationIgnored:   "Ignored",
}

// ---------------- Public API ----------------

// PrintIngressSummary renders the migration report for a single Ingress.
// The output format (table or text) is selected based on the Config.
func (cfg *Config) PrintIngressSummary(ingressReport configs.IngressReport) error {
	if cfg.Table {
		return cfg.printIngressReportTable(ingressReport)
	}

	cfg.printIngressReport(ingressReport)

	return nil
}

// PrintGlobalSummary renders the aggregated migration summary across all Ingresses.
// The output format (table or text) is selected based on the Config.
func (cfg *Config) PrintGlobalSummary(globalReport configs.GlobalReport) error {
	if cfg.Table {
		return cfg.printGlobalSummaryTable(globalReport)
	}

	cfg.printGlobalSummary(globalReport)

	return nil
}

// New returns a new Config with default settings.
func New() *Config {
	return &Config{}
}

// ---------------- Table Renderers ----------------

// printIngressReportTable renders a single Ingress report in table format,
// including a detailed per-annotation table and a summary table.
func (cfg *Config) printIngressReportTable(ingressReport configs.IngressReport) error {
	fmt.Printf("\n=========================== Ingress Report ===========================\n")
	fmt.Printf("\nIngress: %s/%s\n\n", ingressReport.Namespace, ingressReport.Name)

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Annotation", "Status", "Message"})

	rows := make([][]string, 0, len(ingressReport.Entries))

	for _, entries := range ingressReport.Entries {
		msg := entries.Message
		if msg == "" {
			msg = "-"
		}

		rows = append(rows, []string{entries.Name, statusLabel[entries.Status], msg})
	}

	if err := table.Bulk(rows); err != nil {
		return err
	}

	if err := table.Render(); err != nil {
		return err
	}

	// Render per-Ingress summary table.
	summarizedIngress := summarizeIngress(ingressReport)

	return renderSummaryTable("SUMMARY", summarizedIngress)
}

// printGlobalSummaryTable renders the global summary across all Ingresses
// in table format.
func (cfg *Config) printGlobalSummaryTable(globalReport configs.GlobalReport) error {
	fmt.Printf("\n==================================== Global Summary ====================================\n\n")

	total := summarizeGlobal(globalReport)

	return renderSummaryTable("OVERALL RESULT", total)
}

// renderSummaryTable renders a generic summary table given a title and summary counts.
func renderSummaryTable(title string, summaryCounts SummaryCounts) error {
	summary := tablewriter.NewWriter(os.Stdout)
	summary.Header([]string{"Metric", "Count"})

	rows := [][]string{
		{"Converted", strconv.Itoa(summaryCounts.Converted)},
		{"Warnings", strconv.Itoa(summaryCounts.Warnings)},
		{"Skipped", strconv.Itoa(summaryCounts.Skipped)},
		{"Ignored", strconv.Itoa(summaryCounts.Ignored)},
		{"Result", resultLabel(summaryCounts)},
	}

	if err := summary.Bulk(rows); err != nil {
		return err
	}

	fmt.Printf("\n------------------------ %s ------------------------\n", title)

	return summary.Render()
}

// ---------------- Text Renderers ----------------

// printIngressReport renders a single Ingress report in plain text format.
func (cfg *Config) printIngressReport(ingressReport configs.IngressReport) {
	fmt.Printf("\nIngress: %s/%s\n\n", ingressReport.Namespace, ingressReport.Name)

	for _, entries := range ingressReport.Entries {
		switch entries.Status {
		case configs.AnnotationConverted:
			fmt.Printf("  ✅ %s\n", entries.Name)
		case configs.AnnotationWarned:
			fmt.Printf("  ⚠️  %s\n      → %s\n", entries.Name, entries.Message)
		case configs.AnnotationSkipped:
			fmt.Printf("  ❌ %s\n      → %s\n", entries.Name, entries.Message)
		case configs.AnnotationIgnored:
			fmt.Printf("  ℹ️  %s\n", entries.Name)
		}
	}

	summarizedIngress := summarizeIngress(ingressReport)
	printSummaryText(fmt.Sprintf("Summary for %s/%s", ingressReport.Namespace, ingressReport.Name), summarizedIngress)
}

// printGlobalSummary renders the aggregated global summary in plain text format.
func (cfg *Config) printGlobalSummary(globalReport configs.GlobalReport) {
	total := summarizeGlobal(globalReport)

	printSummaryText("Global Summary", total)
}

// printSummaryText prints a human-readable summary block in plain text.
func printSummaryText(title string, summaryCounts SummaryCounts) {
	fmt.Printf("\n================== %s ==================\n", title)
	fmt.Printf("Converted: %d\n", summaryCounts.Converted)
	fmt.Printf("Warnings:  %d\n", summaryCounts.Warnings)
	fmt.Printf("Skipped:   %d\n", summaryCounts.Skipped)
	fmt.Printf("Ignored:   %d\n", summaryCounts.Ignored)
	fmt.Printf("Result:    %s\n\n", resultLabel(summaryCounts))
}

// ---------------- Helpers ----------------

// resultLabel returns a human-readable overall result string based on summary counts.
func resultLabel(summaryCounts SummaryCounts) string {
	if summaryCounts.Skipped > 0 {
		return "❌ Manual action required"
	}

	if summaryCounts.Warnings > 0 {
		return "⚠️ Review recommended"
	}

	return "✅ Clean migration"
}

// summarizeIngress computes summary counts for a single Ingress report.
func summarizeIngress(ingressReport configs.IngressReport) SummaryCounts {
	var summaryCounts SummaryCounts

	for _, entries := range ingressReport.Entries {
		switch entries.Status {
		case configs.AnnotationConverted:
			summaryCounts.Converted++
		case configs.AnnotationWarned:
			summaryCounts.Warnings++
		case configs.AnnotationSkipped:
			summaryCounts.Skipped++
		case configs.AnnotationIgnored:
			summaryCounts.Ignored++
		}
	}

	return summaryCounts
}

// summarizeGlobal computes aggregated summary counts across all Ingress reports.
func summarizeGlobal(globalReport configs.GlobalReport) SummaryCounts {
	var total SummaryCounts

	for _, ir := range globalReport.Ingresses {
		summarizedIngress := summarizeIngress(ir)

		total.Converted += summarizedIngress.Converted
		total.Warnings += summarizedIngress.Warnings
		total.Skipped += summarizedIngress.Skipped
		total.Ignored += summarizedIngress.Ignored
	}

	return total
}
