package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudbees-days/fm-actions-container/internal/cloudbees"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var createFlagCmd = &cobra.Command{
	Use:   "create-flag",
	Short: "Create a new feature flag",
	Long:  `Create a new feature flag with the specified name, type, and configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flagName, _ := cmd.Flags().GetString("flag-name")
		flagType, _ := cmd.Flags().GetString("flag-type")
		description, _ := cmd.Flags().GetString("description")
		variantsStr, _ := cmd.Flags().GetString("variants")
		isPermanent, _ := cmd.Flags().GetBool("is-permanent")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if flagName == "" {
			return fmt.Errorf("flag-name is required")
		}
		if flagType == "" {
			return fmt.Errorf("flag-type is required")
		}

		// Parse variants - try YAML first, fallback to comma-separated
		var variants []string
		if variantsStr != "" {
			// Try parsing as YAML array first
			var yamlVariants []interface{}
			if err := yaml.Unmarshal([]byte(variantsStr), &yamlVariants); err == nil {
				// Successfully parsed as YAML array
				for _, v := range yamlVariants {
					variants = append(variants, fmt.Sprintf("%v", v))
				}
			} else {
				// Fallback to comma-separated parsing
				variants = strings.Split(variantsStr, ",")
				for i, v := range variants {
					variants[i] = strings.TrimSpace(v)
				}
			}
		} else {
			// Default variants based on flag type
			switch strings.ToLower(flagType) {
			case "boolean":
				variants = []string{"true", "false"}
			case "string":
				variants = []string{"option1", "option2"}
			case "number":
				variants = []string{"0", "1"}
			default:
				variants = []string{"true", "false"}
			}
		}

		if dryRun {
			fmt.Printf("DRY RUN: Would create flag '%s'\n", flagName)
			fmt.Printf("Type: %s\n", flagType)
			fmt.Printf("Description: %s\n", description)
			fmt.Printf("Variants: %s\n", strings.Join(variants, ", "))
			fmt.Printf("Permanent: %t\n", isPermanent)
			return nil
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

		flag, err := client.CreateFlag(application.ID, flagName, flagType, description, variants, isPermanent)
		if err != nil {
			return fmt.Errorf("failed to create flag: %w", err)
		}

		// Output results
		flagJSON, _ := json.Marshal(flag)
		cloudbees.WriteOutput("flag-id", flag.ID)
		cloudbees.WriteOutput("flag-name", flag.Name)
		cloudbees.WriteOutput("flag-type", flag.FlagType)
		cloudbees.WriteOutput("flag", string(flagJSON))
		cloudbees.WriteOutput("success", "true")

		if verbose {
			fmt.Printf("Successfully created flag: %s (ID: %s)\n", flag.Name, flag.ID)
			fmt.Printf("Type: %s\n", flag.FlagType)
			if flag.Description != "" {
				fmt.Printf("Description: %s\n", flag.Description)
			}
			fmt.Printf("Variants: %s\n", strings.Join(flag.Variants, ", "))
			fmt.Printf("Permanent: %t\n", flag.IsPermanent)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(createFlagCmd)

	createFlagCmd.Flags().StringP("flag-name", "f", "", "Name of the flag to create (required)")
	createFlagCmd.Flags().StringP("flag-type", "t", "Boolean", "Type of the flag (Boolean, String, Number)")
	createFlagCmd.Flags().StringP("description", "d", "", "Description of the flag")
	createFlagCmd.Flags().String("variants", "", "Variants as YAML array or comma-separated list (defaults based on type)")
	createFlagCmd.Flags().Bool("is-permanent", false, "Whether the flag is permanent")
	createFlagCmd.Flags().Bool("dry-run", false, "Validate flag details without creating")

	createFlagCmd.MarkFlagRequired("flag-name")
	createFlagCmd.MarkPersistentFlagRequired("application-name")
}
