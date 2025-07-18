package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cloudbees-days/fm-actions-container/internal/cloudbees"
	"github.com/spf13/cobra"
)

var getFlagConfigCmd = &cobra.Command{
	Use:   "get-flag-config",
	Short: "Get feature flag configuration",
	Long:  `Get the current feature flag configuration for a specific flag in a given environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagName, _ := cmd.Flags().GetString("flag-name")
		environmentName, _ := cmd.Flags().GetString("environment-name")

		if flagName == "" {
			return fmt.Errorf("flag-name is required")
		}
		if environmentName == "" {
			return fmt.Errorf("environment-name is required")
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

		// Get the flag to retrieve its ID
		flag, err := client.GetFlagByName(application.ID, flagName)
		if err != nil {
			return fmt.Errorf("failed to get flag '%s': %w", flagName, err)
		}

		// Get all environments to find the one that matches the name
		environments, err := client.ListEnvironments()
		if err != nil {
			return fmt.Errorf("failed to list environments: %w", err)
		}

		var environmentID string
		for _, env := range environments {
			if env.Name == environmentName {
				environmentID = env.ID
				break
			}
		}

		if environmentID == "" {
			return fmt.Errorf("environment '%s' not found", environmentName)
		}

		// Get flag configuration
		config, err := client.GetFlagConfiguration(application.ID, flag.ID, environmentID)
		if err != nil {
			return fmt.Errorf("failed to get flag configuration: %w", err)
		}

		// Output results
		configJSON, _ := json.Marshal(config)
		cloudbees.WriteOutput("flag-config", string(configJSON))
		cloudbees.WriteOutput("flag-id", flag.ID)
		cloudbees.WriteOutput("environment-id", environmentID)
		cloudbees.WriteOutput("enabled", fmt.Sprintf("%t", config.Configuration.Enabled))

		// Output default-value as JSON string
		if config.Configuration.DefaultValue != nil {
			defaultValueJSON, _ := json.Marshal(config.Configuration.DefaultValue)
			cloudbees.WriteOutput("default-value", string(defaultValueJSON))
		} else {
			cloudbees.WriteOutput("default-value", "null")
		}

		if verbose {
			fmt.Printf("Flag: %s (ID: %s)\n", flag.Name, flag.ID)
			fmt.Printf("Environment: %s (ID: %s)\n", environmentName, environmentID)
			fmt.Printf("Enabled: %t\n", config.Configuration.Enabled)
			if config.Configuration.DefaultValue != nil {
				defaultValueJSON, _ := json.Marshal(config.Configuration.DefaultValue)
				fmt.Printf("Default Value: %s\n", string(defaultValueJSON))
			}
			fmt.Printf("Variants Enabled: %t\n", config.Configuration.VariantsEnabled)
			if config.Configuration.StickinessProperty != "" {
				fmt.Printf("Stickiness Property: %s\n", config.Configuration.StickinessProperty)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getFlagConfigCmd)

	getFlagConfigCmd.Flags().StringP("flag-name", "f", "", "Flag name (required)")
	getFlagConfigCmd.Flags().StringP("environment-name", "e", "", "Environment name (required)")

	getFlagConfigCmd.MarkFlagRequired("flag-name")
	getFlagConfigCmd.MarkFlagRequired("environment-name")
	getFlagConfigCmd.MarkPersistentFlagRequired("application-name")
}
