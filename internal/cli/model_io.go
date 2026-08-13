package cli

import (
	"context"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/ZsTs119/patchxnote-agent/internal/api"
	"github.com/ZsTs119/patchxnote-agent/internal/modelio"
	"github.com/spf13/cobra"
)

type modelIOLookupFlags struct {
	memoryID  string
	requestID string
	platform  string
	outFile   string
	force     bool
}

type modelIOListFlags struct {
	platform    string
	requestID   string
	taskType    string
	state       string
	recordingID string
	eventID     string
	businessID  string
	dateFrom    string
	dateTo      string
	limit       int
	cursor      string
}

func newModelIOCommand(state *rootState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "model-io",
		Short: "Inspect explicit PatchXNote Agent model IO fields",
	}
	cmd.AddCommand(
		newModelIOListCommand(state),
		newModelIOFieldCommand(state, modelio.FieldSourceText, "source-text", "Print or export the source text/safe transcript projection"),
		newModelIOFieldCommand(state, modelio.FieldProviderResponse, "provider-response", "Print or export the model provider response JSON"),
		newModelIOFieldCommand(state, modelio.FieldParsedResult, "parsed-result", "Print or export the parsed model result JSON"),
		newModelIOFieldCommand(state, modelio.FieldPackagedResult, "packaged-result", "Print or export the packaged structured result JSON"),
		newModelIOExportCommand(state),
	)
	return cmd
}

func newModelIOListCommand(state *rootState) *cobra.Command {
	flags := &modelIOListFlags{limit: 20}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List model IO trace metadata and request IDs",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			params := api.ListModelIOTracesParams{
				Platform: flags.platform, RequestID: flags.requestID, TaskType: flags.taskType, State: flags.state,
				RecordingID: flags.recordingID, EventID: flags.eventID, BusinessID: flags.businessID,
				DateFrom: flags.dateFrom, DateTo: flags.dateTo, Limit: flags.limit, Cursor: flags.cursor,
			}
			if err := validateModelIOListFlags(params); err != nil {
				return err
			}
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			page, err := fetchModelIOTracePage(cmd.Context(), runtime, params)
			if err != nil {
				return err
			}
			return writeModelIOTraceList(cmd, state, page)
		},
	}
	cmd.Flags().StringVar(&flags.platform, "platform", "", "Required platform: mobile or desktop")
	cmd.Flags().StringVar(&flags.requestID, "request-id", "", "Filter by model request/run ID")
	cmd.Flags().StringVar(&flags.taskType, "task-type", "", "Filter by model task type")
	cmd.Flags().StringVar(&flags.state, "state", "", "Filter by trace state")
	cmd.Flags().StringVar(&flags.recordingID, "recording-id", "", "Filter by recording ID")
	cmd.Flags().StringVar(&flags.eventID, "event-id", "", "Filter by event ID")
	cmd.Flags().StringVar(&flags.businessID, "business-id", "", "Filter by business ID")
	cmd.Flags().StringVar(&flags.dateFrom, "date-from", "", "Filter created_at >= RFC3339 timestamp")
	cmd.Flags().StringVar(&flags.dateTo, "date-to", "", "Filter created_at < RFC3339 timestamp")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "Page size, 1-50")
	cmd.Flags().StringVar(&flags.cursor, "cursor", "", "Opaque cursor returned by a previous list call")
	_ = cmd.MarkFlagRequired("platform")
	return cmd
}

func newModelIOFieldCommand(state *rootState, field modelio.Field, use string, short string) *cobra.Command {
	flags := &modelIOLookupFlags{}
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lookup := modelio.Lookup{MemoryID: flags.memoryID, RequestID: flags.requestID, Platform: flags.platform}
			if err := modelio.ValidateLookup(lookup); err != nil {
				return err
			}
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			export, err := fetchModelIOExport(cmd.Context(), runtime, lookup)
			if err != nil {
				return err
			}
			result, err := modelio.SelectField(export, lookup, field)
			if err != nil {
				return err
			}
			if flags.outFile != "" {
				result, err = modelio.WriteFieldFile(flags.outFile, result, flags.force)
				if err != nil {
					return err
				}
				return writeModelIOFieldSummary(cmd, state, result)
			}
			return writeModelIOFieldInline(cmd, state, result)
		},
	}
	addModelIOLookupFlags(cmd, flags, false)
	return cmd
}

func newModelIOExportCommand(state *rootState) *cobra.Command {
	flags := &modelIOLookupFlags{}
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export complete explicit Agent model IO JSON to a local file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lookup := modelio.Lookup{MemoryID: flags.memoryID, RequestID: flags.requestID, Platform: flags.platform}
			if err := modelio.ValidateLookup(lookup); err != nil {
				return err
			}
			if flags.outFile == "" {
				return fmt.Errorf("--out is required")
			}
			runtime, err := loadRuntime(state)
			if err != nil {
				return err
			}
			export, err := fetchModelIOExport(cmd.Context(), runtime, lookup)
			if err != nil {
				return err
			}
			out, err := modelio.WriteExportFile(flags.outFile, export, flags.force)
			if err != nil {
				return err
			}
			switch format := normalizedOutputFormat(state); format {
			case "", "plain":
				_, err = fmt.Fprintf(cmd.OutOrStdout(), "model io exported\nout %s\n", out)
				return err
			case "json":
				return writeJSON(cmd.OutOrStdout(), map[string]any{
					"exported":    true,
					"out":         out,
					"lookup_kind": lookup.Kind(),
					"memory_id":   lookup.MemoryID,
					"request_id":  lookup.RequestID,
					"platform":    outputPlatform(lookup, export),
				})
			default:
				return unsupportedOutputFormatError(format)
			}
		},
	}
	addModelIOLookupFlags(cmd, flags, true)
	return cmd
}

func addModelIOLookupFlags(cmd *cobra.Command, flags *modelIOLookupFlags, requireOut bool) {
	cmd.Flags().StringVar(&flags.memoryID, "memory-id", "", "PatchXNote memory ID")
	cmd.Flags().StringVar(&flags.requestID, "request-id", "", "PatchXNote model request/run ID")
	cmd.Flags().StringVar(&flags.platform, "platform", "", "Optional platform: mobile or desktop")
	cmd.Flags().StringVar(&flags.outFile, "out", "", "Output file")
	cmd.Flags().BoolVar(&flags.force, "force", false, "Overwrite existing output file")
	if requireOut {
		_ = cmd.MarkFlagRequired("out")
	}
}

func fetchModelIOExport(ctx context.Context, runtime runtimeState, lookup modelio.Lookup) (api.AgentModelIOExport, error) {
	if lookup.MemoryID != "" {
		return fetchMemoryModelIO(ctx, runtime, lookup.Platform, lookup.MemoryID)
	}
	return fetchModelRunIOTrace(ctx, runtime, lookup.Platform, lookup.RequestID)
}

func fetchModelIOTracePage(ctx context.Context, runtime runtimeState, params api.ListModelIOTracesParams) (api.AgentModelIOTracePage, error) {
	var page api.AgentModelIOTracePage
	err := withAgentAccessToken(ctx, runtime, func(accessToken string) error {
		var err error
		page, err = runtime.API.ListModelIOTraces(ctx, accessToken, params)
		return err
	})
	return page, err
}

func validateModelIOListFlags(params api.ListModelIOTracesParams) error {
	if params.Platform != "mobile" && params.Platform != "desktop" {
		return fmt.Errorf("--platform must be mobile or desktop")
	}
	if params.Limit < 1 || params.Limit > 50 {
		return fmt.Errorf("--limit must be between 1 and 50")
	}
	var dateFrom time.Time
	var dateTo time.Time
	if params.DateFrom != "" {
		parsed, err := time.Parse(time.RFC3339Nano, params.DateFrom)
		if err != nil {
			return fmt.Errorf("--date-from must be RFC3339")
		}
		dateFrom = parsed
	}
	if params.DateTo != "" {
		parsed, err := time.Parse(time.RFC3339Nano, params.DateTo)
		if err != nil {
			return fmt.Errorf("--date-to must be RFC3339")
		}
		dateTo = parsed
	}
	if !dateFrom.IsZero() && !dateTo.IsZero() && !dateFrom.Before(dateTo) {
		return fmt.Errorf("--date-from must be before --date-to")
	}
	return nil
}

func writeModelIOTraceList(cmd *cobra.Command, state *rootState, page api.AgentModelIOTracePage) error {
	switch format := normalizedOutputFormat(state); format {
	case "", "plain":
		if len(page.Items) == 0 {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), "no model io traces"); err != nil {
				return err
			}
			if page.NextCursor != "" {
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "next_cursor %s\n", page.NextCursor)
				return err
			}
			return nil
		}
		writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(writer, "REQUEST_ID\tPLATFORM\tTASK_TYPE\tSTATE\tSOURCE\tPROVIDER\tPARSED\tPACKAGED\tCREATED_AT\tCOMPLETED_AT\tMEMORY")
		for _, item := range page.Items {
			memory := "no"
			if item.Memory != nil && item.Memory.ID != "" {
				memory = item.Memory.ID
			}
			fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				item.RequestID, item.Platform, item.TaskType, item.State, item.SourceTextAvailability,
				item.FieldStatus.ProviderResponseJSON, item.FieldStatus.ParsedResultJSON, item.FieldStatus.PackagedResultJSON,
				formatModelIOListTime(item.CreatedAt), formatModelIOListTimePtr(item.CompletedAt), memory)
		}
		if err := writer.Flush(); err != nil {
			return err
		}
		if page.NextCursor != "" {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "next_cursor %s\n", page.NextCursor)
			return err
		}
		return nil
	case "json":
		return writeJSON(cmd.OutOrStdout(), page)
	default:
		return unsupportedOutputFormatError(format)
	}
}

func formatModelIOListTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func formatModelIOListTimePtr(value *time.Time) string {
	if value == nil {
		return ""
	}
	return formatModelIOListTime(*value)
}

func writeModelIOFieldInline(cmd *cobra.Command, state *rootState, result modelio.FieldResult) error {
	switch format := normalizedOutputFormat(state); format {
	case "", "plain":
		if !result.Available {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "model io field unavailable\nfield %s\nstatus %s\n", result.Field, result.Status)
			return err
		}
		body, err := result.ContentBytes()
		if err != nil {
			return err
		}
		if len(body) > modelio.DefaultInlineLimitBytes {
			return fmt.Errorf("model IO field exceeds inline output limit; pass --out to write it to a file")
		}
		_, err = cmd.OutOrStdout().Write(body)
		return err
	case "json":
		includeContent := result.Available
		if includeContent {
			body, err := result.ContentBytes()
			if err != nil {
				return err
			}
			if len(body) > modelio.DefaultInlineLimitBytes {
				return fmt.Errorf("model IO field exceeds inline output limit; pass --out to write it to a file")
			}
		}
		summary, err := result.Summary(includeContent)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), summary)
	default:
		return unsupportedOutputFormatError(format)
	}
}

func writeModelIOFieldSummary(cmd *cobra.Command, state *rootState, result modelio.FieldResult) error {
	switch format := normalizedOutputFormat(state); format {
	case "", "plain":
		if !result.Available {
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "model io field unavailable\nfield %s\nstatus %s\n", result.Field, result.Status)
			return err
		}
		_, err := fmt.Fprintf(cmd.OutOrStdout(), "model io field exported\nfield %s\nout %s\n", result.Field, result.OutputPath)
		return err
	case "json":
		summary, err := result.Summary(false)
		if err != nil {
			return err
		}
		return writeJSON(cmd.OutOrStdout(), summary)
	default:
		return unsupportedOutputFormatError(format)
	}
}

func outputPlatform(lookup modelio.Lookup, export api.AgentModelIOExport) string {
	if lookup.Platform != "" {
		return lookup.Platform
	}
	if export.Memory != nil && export.Memory.Platform != "" {
		return export.Memory.Platform
	}
	return export.Trace.Platform
}
