package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mova-engine/mova-engine/core/executor"
	"github.com/spf13/cobra"
)

// NewDLQCommand creates the dlq command with subcommands
func NewDLQCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dlq",
		Short: "Manage Dead Letter Queue entries",
		Long:  `Manage failed workflows in the Dead Letter Queue (DLQ)`,
	}

	// Add subcommands
	cmd.AddCommand(newDLQListCommand())
	cmd.AddCommand(newDLQShowCommand())
	cmd.AddCommand(newDLQRetryCommand())
	cmd.AddCommand(newDLQArchiveCommand())
	cmd.AddCommand(newDLQDeleteCommand())
	cmd.AddCommand(newDLQStatsCommand())

	return cmd
}

// newDLQListCommand creates the dlq list command
func newDLQListCommand() *cobra.Command {
	var (
		status       string
		workflowType string
		userID       string
		limit        int
		format       string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DLQ entries",
		Long:  `List all entries in the Dead Letter Queue with optional filtering`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dlqPath := getDLQPath()
			dlq := executor.NewDeadLetterQueue(dlqPath)

			filter := executor.DLQFilter{
				WorkflowType: workflowType,
				UserID:       userID,
				Limit:        limit,
			}

			if status != "" {
				dlqStatus := executor.DLQStatus(status)
				filter.Status = &dlqStatus
			}

			entries, err := dlq.List(filter)
			if err != nil {
				return fmt.Errorf("failed to list DLQ entries: %w", err)
			}

			if len(entries) == 0 {
				fmt.Println("No DLQ entries found")
				return nil
			}

			switch format {
			case "json":
				return printDLQEntriesJSON(entries)
			case "table":
				return printDLQEntriesTable(entries)
			default:
				return printDLQEntriesTable(entries)
			}
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (active, retrying, resolved, archived)")
	cmd.Flags().StringVar(&workflowType, "workflow-type", "", "Filter by workflow type")
	cmd.Flags().StringVar(&userID, "user-id", "", "Filter by user ID")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of results")
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json)")

	return cmd
}

// newDLQShowCommand creates the dlq show command
func newDLQShowCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "show <dlq-id>",
		Short: "Show detailed information about a DLQ entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dlqID := args[0]
			dlqPath := getDLQPath()
			dlq := executor.NewDeadLetterQueue(dlqPath)

			entry, err := dlq.Get(dlqID)
			if err != nil {
				return fmt.Errorf("failed to get DLQ entry: %w", err)
			}

			switch format {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(entry)
			default:
				return printDLQEntryDetailed(entry)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "detailed", "Output format (detailed, json)")

	return cmd
}

// newDLQRetryCommand creates the dlq retry command
func newDLQRetryCommand() *cobra.Command {
	var (
		sandboxMode bool
		wait        bool
	)

	cmd := &cobra.Command{
		Use:   "retry <dlq-id>",
		Short: "Retry a failed workflow from DLQ",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dlqID := args[0]
			dlqPath := getDLQPath()
			dlq := executor.NewDeadLetterQueue(dlqPath)

			// Get the DLQ entry first to show info
			entry, err := dlq.Get(dlqID)
			if err != nil {
				return fmt.Errorf("failed to get DLQ entry: %w", err)
			}

			fmt.Printf("Retrying DLQ entry: %s\n", dlqID)
			fmt.Printf("Original Run ID: %s\n", entry.RunID)
			fmt.Printf("Workflow Type: %s\n", entry.Metadata.WorkflowType)
			fmt.Printf("Sandbox Mode: %v\n", sandboxMode)
			fmt.Println()

			// For CLI, we'll simulate the retry by updating status
			// In a real implementation, this would call the API
			err = dlq.UpdateStatus(dlqID, executor.DLQStatusRetrying)
			if err != nil {
				return fmt.Errorf("failed to update DLQ status: %w", err)
			}

			fmt.Printf("✓ DLQ entry %s marked for retry\n", dlqID)
			fmt.Println("Note: Use the API endpoint for actual workflow execution")

			return nil
		},
	}

	cmd.Flags().BoolVar(&sandboxMode, "sandbox", true, "Run in sandbox mode (isolated execution)")
	cmd.Flags().BoolVar(&wait, "wait", false, "Wait for execution to complete")

	return cmd
}

// newDLQArchiveCommand creates the dlq archive command
func newDLQArchiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive <dlq-id>",
		Short: "Archive a DLQ entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dlqID := args[0]
			dlqPath := getDLQPath()
			dlq := executor.NewDeadLetterQueue(dlqPath)

			err := dlq.Archive(dlqID)
			if err != nil {
				return fmt.Errorf("failed to archive DLQ entry: %w", err)
			}

			fmt.Printf("✓ DLQ entry %s archived successfully\n", dlqID)
			return nil
		},
	}

	return cmd
}

// newDLQDeleteCommand creates the dlq delete command
func newDLQDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <dlq-id>",
		Short: "Delete a DLQ entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dlqID := args[0]

			if !force {
				fmt.Printf("Are you sure you want to delete DLQ entry %s? (y/N): ", dlqID)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("Delete cancelled")
					return nil
				}
			}

			dlqPath := getDLQPath()
			dlq := executor.NewDeadLetterQueue(dlqPath)

			err := dlq.Delete(dlqID)
			if err != nil {
				return fmt.Errorf("failed to delete DLQ entry: %w", err)
			}

			fmt.Printf("✓ DLQ entry %s deleted successfully\n", dlqID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force delete without confirmation")

	return cmd
}

// newDLQStatsCommand creates the dlq stats command
func newDLQStatsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show DLQ statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			dlqPath := getDLQPath()
			dlq := executor.NewDeadLetterQueue(dlqPath)

			entries, err := dlq.List(executor.DLQFilter{})
			if err != nil {
				return fmt.Errorf("failed to get DLQ entries: %w", err)
			}

			return printDLQStats(entries)
		},
	}

	return cmd
}

// Helper functions

func getDLQPath() string {
	if path := os.Getenv("MOVA_DLQ_PATH"); path != "" {
		return path
	}
	return "./state/deadletter"
}

func printDLQEntriesTable(entries []*executor.DeadLetterEntry) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "ID\tRUN ID\tWORKFLOW TYPE\tSTATUS\tCREATED\tATTEMPTS\tLAST ERROR")
	fmt.Fprintln(w, "--\t------\t-------------\t------\t-------\t--------\t----------")

	for _, entry := range entries {
		createdAt := entry.CreatedAt.Format("2006-01-02 15:04")
		lastError := entry.ErrorDetails.LastError
		if len(lastError) > 50 {
			lastError = lastError[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			entry.ID[:8],
			entry.RunID[:12],
			entry.Metadata.WorkflowType,
			entry.Status,
			createdAt,
			entry.ErrorDetails.Attempts,
			lastError,
		)
	}

	return nil
}

func printDLQEntriesJSON(entries []*executor.DeadLetterEntry) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

func printDLQEntryDetailed(entry *executor.DeadLetterEntry) error {
	fmt.Printf("DLQ Entry Details\n")
	fmt.Printf("================\n\n")

	fmt.Printf("ID:               %s\n", entry.ID)
	fmt.Printf("Run ID:           %s\n", entry.RunID)
	fmt.Printf("Status:           %s\n", entry.Status)
	fmt.Printf("Created:          %s\n", entry.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Workflow Type:    %s\n", entry.Metadata.WorkflowType)
	fmt.Printf("Retry Count:      %d\n", entry.Metadata.RetryCount)

	if entry.Metadata.LastRetryAt != nil {
		fmt.Printf("Last Retry:       %s\n", entry.Metadata.LastRetryAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("\nError Details:\n")
	fmt.Printf("  Attempts:       %d\n", entry.ErrorDetails.Attempts)
	fmt.Printf("  Failure Reason: %s\n", entry.ErrorDetails.FailureReason)
	fmt.Printf("  Last Error:     %s\n", entry.ErrorDetails.LastError)

	if len(entry.ErrorDetails.ErrorHistory) > 0 {
		fmt.Printf("\nError History:\n")
		for i, err := range entry.ErrorDetails.ErrorHistory {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
	}

	if entry.FailedAction != nil {
		fmt.Printf("\nFailed Action:\n")
		fmt.Printf("  Name: %s\n", entry.FailedAction.Name)
		fmt.Printf("  Type: %s\n", entry.FailedAction.Type)
	}

	fmt.Printf("\nWorkflow Intent:\n")
	fmt.Printf("  Name:        %s\n", entry.Envelope.Intent.Name)
	fmt.Printf("  Version:     %s\n", entry.Envelope.Intent.Version)
	fmt.Printf("  Description: %s\n", entry.Envelope.Intent.Description)

	if len(entry.Envelope.Intent.Tags) > 0 {
		fmt.Printf("  Tags:        %s\n", strings.Join(entry.Envelope.Intent.Tags, ", "))
	}

	return nil
}

func printDLQStats(entries []*executor.DeadLetterEntry) error {
	if len(entries) == 0 {
		fmt.Println("No DLQ entries found")
		return nil
	}

	// Calculate statistics
	statusCounts := make(map[executor.DLQStatus]int)
	workflowTypeCounts := make(map[string]int)
	var oldestEntry, newestEntry *executor.DeadLetterEntry

	for _, entry := range entries {
		statusCounts[entry.Status]++
		workflowTypeCounts[entry.Metadata.WorkflowType]++

		if oldestEntry == nil || entry.CreatedAt.Before(oldestEntry.CreatedAt) {
			oldestEntry = entry
		}
		if newestEntry == nil || entry.CreatedAt.After(newestEntry.CreatedAt) {
			newestEntry = entry
		}
	}

	fmt.Printf("DLQ Statistics\n")
	fmt.Printf("==============\n\n")

	fmt.Printf("Total Entries: %d\n\n", len(entries))

	fmt.Printf("By Status:\n")
	for status, count := range statusCounts {
		fmt.Printf("  %-10s: %d\n", status, count)
	}

	fmt.Printf("\nBy Workflow Type:\n")
	for workflowType, count := range workflowTypeCounts {
		fmt.Printf("  %-20s: %d\n", workflowType, count)
	}

	if oldestEntry != nil {
		fmt.Printf("\nOldest Entry: %s (%s)\n",
			oldestEntry.CreatedAt.Format("2006-01-02 15:04:05"),
			oldestEntry.ID[:8])
	}

	if newestEntry != nil {
		fmt.Printf("Newest Entry: %s (%s)\n",
			newestEntry.CreatedAt.Format("2006-01-02 15:04:05"),
			newestEntry.ID[:8])
	}

	return nil
}
