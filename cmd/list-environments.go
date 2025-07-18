package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cloudbees-days/fm-actions-container/internal/cloudbees"
	"github.com/spf13/cobra"
)

var listEnvironmentsCmd = &cobra.Command{
	Use:   "list-environments",
	Short: "List all environments in the organization",
	Long:  `List all environments in the organization for feature flag targeting and configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get authentication parameters from root command
		apiURL, _ := cmd.Root().PersistentFlags().GetString("api-url")
		token, _ := cmd.Root().PersistentFlags().GetString("token")
		orgID, _ := cmd.Root().PersistentFlags().GetString("org-id")

		client, err := cloudbees.NewClient(apiURL, token, orgID)
		if err != nil {
			return fmt.Errorf("failed to create CloudBees client: %w", err)
		}

		environments, err := client.ListEnvironments()
		if err != nil {
			return fmt.Errorf("failed to list environments: %w", err)
		}

		if len(environments) == 0 {
			fmt.Println("No environments found")
			cloudbees.WriteOutput("environment-count", "0")
			cloudbees.WriteOutput("environments", "[]")
			return nil
		}

		// Output results
		environmentsJSON, _ := json.Marshal(environments)
		cloudbees.WriteOutput("environment-count", fmt.Sprintf("%d", len(environments)))
		cloudbees.WriteOutput("environments", string(environmentsJSON))

		if verbose {
			fmt.Printf("Found %d environments:\n", len(environments))
			for _, env := range environments {
				status := "active"
				if env.IsDisabled {
					status = "disabled"
				}
				fmt.Printf("- %s (ID: %s, Status: %s)\n", env.Name, env.ID, status)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listEnvironmentsCmd)
}
