package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cloudbees-days/fm-actions-container/internal/cloudbees"
	"github.com/spf13/cobra"
)

var listFlagsCmd = &cobra.Command{
	Use:   "list-flags",
	Short: "List all feature flags in the organization",
	Long:  `List all feature flags in the organization with their metadata and current status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get authentication parameters from root command
		apiURL, _ := cmd.Root().PersistentFlags().GetString("api-url")
		token, _ := cmd.Root().PersistentFlags().GetString("token")
		orgID, _ := cmd.Root().PersistentFlags().GetString("org-id")
		applicationName, _ := cmd.Root().PersistentFlags().GetString("application-name")

		client, err := cloudbees.NewClient(apiURL, token, orgID)
		if err != nil {
			return fmt.Errorf("failed to create CloudBees client: %w", err)
		}

		// First, get the application to retrieve its ID
		application, err := client.GetApplicationByName(applicationName)
		if err != nil {
			return fmt.Errorf("failed to get application '%s': %w", applicationName, err)
		}

		flags, err := client.ListFlags(application.ID)
		if err != nil {
			return fmt.Errorf("failed to list flags: %w", err)
		}

		if len(flags) == 0 {
			fmt.Println("No flags found")
			cloudbees.WriteOutput("flag-count", "0")
			cloudbees.WriteOutput("flags", "[]")
			return nil
		}

		// Output results
		flagsJSON, _ := json.Marshal(flags)
		cloudbees.WriteOutput("flag-count", fmt.Sprintf("%d", len(flags)))
		cloudbees.WriteOutput("flags", string(flagsJSON))

		if verbose {
			fmt.Printf("Found %d flags:\n", len(flags))
			for _, flag := range flags {
				permanent := "temporary"
				if flag.IsPermanent {
					permanent = "permanent"
				}
				fmt.Printf("- %s (ID: %s, Type: %s, %s)\n", flag.Name, flag.ID, flag.FlagType, permanent)
				if flag.Description != "" {
					fmt.Printf("  Description: %s\n", flag.Description)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listFlagsCmd)
	listFlagsCmd.MarkPersistentFlagRequired("application-name")
}
