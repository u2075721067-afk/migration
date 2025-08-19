package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/mova-engine/mova-engine/cli/commands"
	"github.com/mova-engine/mova-engine/core/validator"
	"github.com/spf13/cobra"
)

var (
	apiServer string
	wait      bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "mova",
		Short: "MOVA Automation Engine CLI",
		Long: `MOVA (Model-View-Automation) Engine CLI tool for executing and managing workflows.
		
This tool allows you to:
- Execute MOVA workflow envelopes
- Validate workflow definitions
- View execution logs and status
- Manage workflow runs`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiServer, "server", "http://localhost:8080", "API server URL")

	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(commands.NewDLQCommand())

	// Add config commands
	rootCmd.AddCommand(commands.ConfigCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var runCmd = &cobra.Command{
	Use:   "run [workflow-file]",
	Short: "Execute a MOVA workflow",
	Long:  `Execute a MOVA workflow from a JSON file. The file should contain a valid MOVA v3.1 envelope.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		// Read workflow file
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		// Parse JSON
		var envelope map[string]interface{}
		if err := json.Unmarshal(data, &envelope); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		// Send to API server for execution
		url := fmt.Sprintf("%s/v1/execute", apiServer)
		if wait {
			url += "?wait=true"
		}

		reqBody, err := json.Marshal(envelope)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling envelope: %v\n", err)
			os.Exit(1)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error sending request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
			os.Exit(1)
		}

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
			var result map[string]interface{}
			if err := json.Unmarshal(body, &result); err != nil {
				fmt.Printf("Response: %s\n", string(body))
			} else {
				if wait {
					fmt.Println("✅ Workflow executed successfully")
					fmt.Printf("Result: %+v\n", result)
				} else {
					fmt.Println("✅ Workflow execution started")
					fmt.Printf("Run ID: %v\n", result["run_id"])
				}
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Status)
			fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
			os.Exit(1)
		}
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [workflow-file]",
	Short: "Validate a MOVA workflow",
	Long:  `Validate a MOVA workflow file against the v3.1 specification.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		// Get schema path (relative to current working directory)
		schemaPath := "./schemas"

		// Create validator
		v, err := validator.NewValidator(schemaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating validator: %v\n", err)
			os.Exit(1)
		}

		// Validate envelope
		valid, errors := v.ValidateEnvelope(filename)
		if !valid {
			fmt.Fprintf(os.Stderr, "Validation failed for %s:\n", filename)
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "  - %v\n", err)
			}
			os.Exit(1)
		}

		fmt.Printf("✅ Workflow file %s is valid\n", filename)
		fmt.Println("Schema validation passed successfully")
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs [run-id]",
	Short: "View execution logs",
	Long:  `View logs for a specific workflow execution run.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runID := args[0]

		// Fetch logs from API server
		url := fmt.Sprintf("%s/v1/runs/%s/logs", apiServer, runID)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching logs: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Logs for run %s:\n", runID)
			fmt.Println(string(body))
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Status)
			if err == nil {
				fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
			}
			os.Exit(1)
		}
	},
}

var statusCmd = &cobra.Command{
	Use:   "status [run-id]",
	Short: "View execution status",
	Long:  `View the current status of a workflow execution run.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runID := args[0]

		// Fetch status from API server
		url := fmt.Sprintf("%s/v1/runs/%s", apiServer, runID)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching status: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
				os.Exit(1)
			}

			var status map[string]interface{}
			if err := json.Unmarshal(body, &status); err != nil {
				fmt.Printf("Status for run %s:\n", runID)
				fmt.Println(string(body))
			} else {
				fmt.Printf("Status for run %s:\n", runID)
				fmt.Printf("Status: %v\n", status["status"])
				if status["start_time"] != nil {
					fmt.Printf("Start Time: %v\n", status["start_time"])
				}
				if status["end_time"] != nil {
					fmt.Printf("End Time: %v\n", status["end_time"])
				}
			}
		} else {
			body, err := ioutil.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Status)
			if err == nil {
				fmt.Fprintf(os.Stderr, "Response: %s\n", string(body))
			}
			os.Exit(1)
		}
	},
}

func init() {
	runCmd.Flags().BoolVar(&wait, "wait", false, "Wait for execution to complete")
}
