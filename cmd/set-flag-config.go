package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudbees-days/fm-actions-container/internal/cloudbees"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var setFlagConfigCmd = &cobra.Command{
	Use:   "set-flag-config",
	Short: "Set feature flag configuration",
	Long:  `Set feature flag configuration (enable/disable flag, set default value) for a target environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagName, _ := cmd.Flags().GetString("flag-name")
		environmentName, _ := cmd.Flags().GetString("environment-name")
		enabled, _ := cmd.Flags().GetString("enabled")
		defaultValue, _ := cmd.Flags().GetString("default-value")
		variantsEnabled, _ := cmd.Flags().GetString("variants-enabled")
		stickinessProperty, _ := cmd.Flags().GetString("stickiness-property")
		configYAML, _ := cmd.Flags().GetString("config")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

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

		// Build configuration map with only the fields that were specified
		configChanges := make(map[string]interface{})

		// Parse and apply configuration from YAML if provided
		if configYAML != "" {
			if err := yaml.Unmarshal([]byte(configYAML), &configChanges); err != nil {
				return fmt.Errorf("failed to parse config YAML: %w", err)
			}
		}

		// Apply individual flag overrides (these take precedence over YAML)
		if enabled != "" {
			enabledBool, err := strconv.ParseBool(enabled)
			if err != nil {
				return fmt.Errorf("invalid enabled value '%s', must be true or false", enabled)
			}
			configChanges["enabled"] = enabledBool
		}

		if defaultValue != "" {
			// Try to parse as JSON first, fallback to string
			var parsedValue interface{}
			if err := json.Unmarshal([]byte(defaultValue), &parsedValue); err != nil {
				// If JSON parsing fails, treat as string
				parsedValue = defaultValue
			}
			configChanges["defaultValue"] = parsedValue
		}

		if variantsEnabled != "" {
			variantsBool, err := strconv.ParseBool(variantsEnabled)
			if err != nil {
				return fmt.Errorf("invalid variants-enabled value '%s', must be true or false", variantsEnabled)
			}
			configChanges["variantsEnabled"] = variantsBool
		}

		if stickinessProperty != "" {
			configChanges["stickinessProperty"] = stickinessProperty
		}

		// Ensure we have at least one field to update
		if len(configChanges) == 0 {
			return fmt.Errorf("no configuration changes specified")
		}

		// For dry-run, convert to FlagConfiguration struct for display
		var newConfig cloudbees.FlagConfiguration
		if dryRun {
			configJSON, _ := json.Marshal(configChanges)
			json.Unmarshal(configJSON, &newConfig)
		}

		if dryRun {
			fmt.Printf("DRY RUN: Would update flag '%s' in environment '%s'\n", flagName, environmentName)
			configJSON, _ := json.MarshalIndent(configChanges, "", "  ")
			fmt.Printf("Configuration changes:\n%s\n", configJSON)
			return nil
		}

		// Set flag configuration using PUT with only specified fields
		err = client.SetFlagConfiguration(application.ID, flag.ID, environmentID, configChanges)
		if err != nil {
			return fmt.Errorf("failed to set flag configuration: %w", err)
		}

		// Output results
		configJSON, _ := json.Marshal(configChanges)
		cloudbees.WriteOutput("flag-id", flag.ID)
		cloudbees.WriteOutput("flag-name", flag.Name)
		cloudbees.WriteOutput("application-id", application.ID)
		cloudbees.WriteOutput("application-name", application.Name)
		cloudbees.WriteOutput("environment-id", environmentID)
		cloudbees.WriteOutput("environment-name", environmentName)
		cloudbees.WriteOutput("configuration", string(configJSON))
		if enabled, ok := configChanges["enabled"].(bool); ok {
			cloudbees.WriteOutput("enabled", fmt.Sprintf("%t", enabled))
		}
		cloudbees.WriteOutput("success", "true")

		if verbose {
			fmt.Printf("Successfully updated flag: %s (ID: %s)\n", flag.Name, flag.ID)
			fmt.Printf("Environment: %s (ID: %s)\n", environmentName, environmentID)
			fmt.Printf("Applied changes:\n")
			for key, value := range configChanges {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(setFlagConfigCmd)

	setFlagConfigCmd.Flags().StringP("flag-name", "f", "", "Flag name (required)")
	setFlagConfigCmd.Flags().StringP("environment-name", "e", "", "Environment name (required)")
	setFlagConfigCmd.Flags().String("enabled", "", "Enable/disable the flag (true/false)")
	setFlagConfigCmd.Flags().String("default-value", "", "Default value for the flag (JSON or string)")
	setFlagConfigCmd.Flags().String("variants-enabled", "", "Enable/disable variants (true/false)")
	setFlagConfigCmd.Flags().String("stickiness-property", "", "Stickiness property for consistent evaluation")
	setFlagConfigCmd.Flags().String("config", "", "Complete configuration as YAML")
	setFlagConfigCmd.Flags().Bool("dry-run", false, "Validate configuration without applying changes")

	setFlagConfigCmd.MarkFlagRequired("flag-name")
	setFlagConfigCmd.MarkFlagRequired("environment-name")
	setFlagConfigCmd.MarkPersistentFlagRequired("application-name")
}
