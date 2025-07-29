package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	apiURL  string
	token   string
	orgID   string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fm-actions",
	Short: "CloudBees Feature Management Actions CLI",
	Long: `A unified CLI tool for CloudBees Feature Management actions including:
- Getting feature flag configurations
- Setting feature flag configurations  
- Listing environments
- Managing feature flags across environments`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().String("token", "", "CloudBees Platform API token (required)")
	rootCmd.PersistentFlags().String("org-id", "", "Organization ID (required)")
	rootCmd.PersistentFlags().String("application-name", "", "Application name (required)")
	rootCmd.PersistentFlags().String("api-url", "https://api.cloudbees.io", "CloudBees Platform API URL")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("use-org-as-app", false, "Use organization ID as application ID for flags API (legacy mode)")

	// Mark required flags
	rootCmd.MarkPersistentFlagRequired("token")
	rootCmd.MarkPersistentFlagRequired("org-id")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".fm-actions" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".fm-actions")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
