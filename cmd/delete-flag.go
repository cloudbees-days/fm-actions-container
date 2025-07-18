package cmd

import (
	"fmt"

	"github.com/cloudbees-days/fm-actions-container/internal/cloudbees"
	"github.com/spf13/cobra"
)

var deleteFlagCmd = &cobra.Command{
	Use:   "delete-flag",
	Short: "Delete a feature flag",
	Long:  `Delete a feature flag by name. This action cannot be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagName, _ := cmd.Flags().GetString("flag-name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		confirm, _ := cmd.Flags().GetBool("confirm")

		if flagName == "" {
			return fmt.Errorf("flag-name is required")
		}

		if !confirm && !dryRun {
			return fmt.Errorf("this action will permanently delete the flag. Use --confirm to proceed or --dry-run to preview")
		}

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

		// Get the flag to retrieve its ID and verify it exists
		flag, err := client.GetFlagByName(application.ID, flagName)
		if err != nil {
			return fmt.Errorf("failed to find flag '%s': %w", flagName, err)
		}

		if dryRun {
			fmt.Printf("DRY RUN: Would delete flag '%s' (ID: %s)\n", flag.Name, flag.ID)
			fmt.Printf("Type: %s\n", flag.FlagType)
			if flag.Description != "" {
				fmt.Printf("Description: %s\n", flag.Description)
			}
			fmt.Printf("Permanent: %t\n", flag.IsPermanent)
			return nil
		}

		// Delete the flag
		err = client.DeleteFlag(application.ID, flag.ID)
		if err != nil {
			return fmt.Errorf("failed to delete flag: %w", err)
		}

		// Output results
		cloudbees.WriteOutput("flag-id", flag.ID)
		cloudbees.WriteOutput("flag-name", flag.Name)
		cloudbees.WriteOutput("deleted", "true")
		cloudbees.WriteOutput("success", "true")

		if verbose {
			fmt.Printf("Successfully deleted flag: %s (ID: %s)\n", flag.Name, flag.ID)
		} else {
			fmt.Printf("Flag '%s' deleted successfully\n", flag.Name)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteFlagCmd)

	deleteFlagCmd.Flags().StringP("flag-name", "f", "", "Name of the flag to delete (required)")
	deleteFlagCmd.Flags().Bool("dry-run", false, "Preview the deletion without actually deleting")
	deleteFlagCmd.Flags().Bool("confirm", false, "Confirm that you want to delete the flag (required unless using dry-run)")

	deleteFlagCmd.MarkFlagRequired("flag-name")
	deleteFlagCmd.MarkPersistentFlagRequired("application-name")
}
