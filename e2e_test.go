package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Load .env file if it exists
	godotenv.Load()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// Helper function to run CLI commands
func runCLI(args ...string) (string, error) {
	cmd := exec.Command("./fm-actions", args...)

	// Set environment variables
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Helper function to run CLI commands with CloudBees outputs
func runCLIWithOutputs(args ...string) (string, string, error) {
	// Create temporary directory for outputs
	outputDir, err := ioutil.TempDir("", "cloudbees_outputs_*")
	if err != nil {
		return "", "", err
	}

	cmd := exec.Command("./fm-actions", args...)

	// Set environment variables including CLOUDBEES_OUTPUTS
	env := os.Environ()
	env = append(env, fmt.Sprintf("CLOUDBEES_OUTPUTS=%s", outputDir))
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	return string(output), outputDir, err
}

// Helper function to read CloudBees output file
func readOutput(outputDir, name string) (string, error) {
	filepath := filepath.Join(outputDir, name)
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Helper function to check if CloudBees output file exists
func outputExists(outputDir, name string) bool {
	filepath := filepath.Join(outputDir, name)
	_, err := os.Stat(filepath)
	return err == nil
}

// hasRequiredEnvVars checks if required environment variables are set for E2E tests
func hasRequiredEnvVars(t *testing.T) bool {
	required := []string{"CLOUDBEES_TOKEN", "CLOUDBEES_ORG_ID", "TEST_APPLICATION_NAME", "TEST_ENVIRONMENT_NAME"}

	for _, env := range required {
		if os.Getenv(env) == "" {
			t.Logf("Required environment variable %s is not set", env)
			return false
		}
	}

	return true
}

// TestCLIHelp tests that the CLI help works
func TestCLIHelp(t *testing.T) {
	output, err := runCLI("--help")
	require.NoError(t, err)
	assert.Contains(t, output, "CloudBees Feature Management actions")
	assert.Contains(t, output, "list-environments")
	assert.Contains(t, output, "get-flag-config")
	assert.Contains(t, output, "set-flag-config")
	assert.Contains(t, output, "create-flag")
	assert.Contains(t, output, "delete-flag")
	assert.Contains(t, output, "list-flags")
}

// TestCommandHelp tests that individual command help works
func TestCommandHelp(t *testing.T) {
	commands := []string{"list-environments", "get-flag-config", "set-flag-config", "create-flag", "delete-flag", "list-flags"}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			output, err := runCLI(cmd, "--help")
			assert.NoError(t, err)
			assert.Contains(t, output, "Usage:")
			assert.Contains(t, output, "Flags:")
		})
	}
}

// TestMissingRequiredFlags tests that commands fail with missing required flags
func TestMissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "list-environments missing token",
			args:     []string{"list-environments", "--org-id=test"},
			expected: "required flag(s) \"token\" not set",
		},
		{
			name:     "list-environments missing org-id",
			args:     []string{"list-environments", "--token=test"},
			expected: "required flag(s) \"org-id\" not set",
		},
		{
			name:     "get-flag-config missing flag-name",
			args:     []string{"get-flag-config", "--token=test", "--org-id=test", "--application-name=test", "--environment-name=test"},
			expected: "required flag(s) \"flag-name\" not set",
		},
		{
			name:     "create-flag missing flag-name",
			args:     []string{"create-flag", "--token=test", "--org-id=test", "--application-name=test"},
			expected: "required flag(s) \"flag-name\" not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runCLI(tt.args...)
			require.Error(t, err)
			assert.Contains(t, output, tt.expected)
		})
	}
}

// TestE2EListEnvironments tests the list-environments command with real API
func TestE2EListEnvironments(t *testing.T) {
	if !hasRequiredEnvVars(t) {
		t.Skip("Skipping E2E test - required environment variables not set")
	}

	token := os.Getenv("CLOUDBEES_TOKEN")
	orgID := os.Getenv("CLOUDBEES_ORG_ID")

	output, outputDir, err := runCLIWithOutputs("list-environments",
		"--token", token,
		"--org-id", orgID,
		"--verbose")

	defer os.RemoveAll(outputDir)

	require.NoError(t, err)
	assert.Contains(t, output, "Found")

	// Check CloudBees outputs
	assert.True(t, outputExists(outputDir, "environment-count"))
	assert.True(t, outputExists(outputDir, "environments"))

	envCount, err := readOutput(outputDir, "environment-count")
	require.NoError(t, err)
	assert.NotEqual(t, "0", envCount)
}

// TestE2EListFlags tests the list-flags command
func TestE2EListFlags(t *testing.T) {
	if !hasRequiredEnvVars(t) {
		t.Skip("Skipping E2E test - required environment variables not set")
	}

	token := os.Getenv("CLOUDBEES_TOKEN")
	orgID := os.Getenv("CLOUDBEES_ORG_ID")
	appName := os.Getenv("TEST_APPLICATION_NAME")

	_, outputDir, err := runCLIWithOutputs("list-flags",
		"--token", token,
		"--org-id", orgID,
		"--application-name", appName,
		"--verbose")

	defer os.RemoveAll(outputDir)

	require.NoError(t, err)

	// Check CloudBees outputs
	assert.True(t, outputExists(outputDir, "flag-count"))
	assert.True(t, outputExists(outputDir, "flags"))
}

// createTestFlag creates a test flag for E2E testing with cleanup
func createTestFlag(t *testing.T, token, orgID, appName string) string {
	flagName := fmt.Sprintf("e2e-test-flag-%d", time.Now().Unix())

	output, outputDir, err := runCLIWithOutputs("create-flag",
		"--token", token,
		"--org-id", orgID,
		"--application-name", appName,
		"--flag-name", flagName,
		"--flag-type", "Boolean",
		"--description", "E2E test flag - safe to delete",
		"--verbose")

	defer os.RemoveAll(outputDir)

	require.NoError(t, err, "Failed to create test flag")
	assert.Contains(t, output, "Successfully created flag")

	// Verify outputs
	assert.True(t, outputExists(outputDir, "flag-id"))
	assert.True(t, outputExists(outputDir, "flag-name"))

	flagID, err := readOutput(outputDir, "flag-id")
	require.NoError(t, err)
	require.NotEmpty(t, flagID)

	// Set up cleanup
	t.Cleanup(func() {
		// First disable the flag (enabled flags can't be deleted)
		envName := os.Getenv("TEST_ENVIRONMENT_NAME")
		if envName != "" {
			disableConfig := `enabled: false`
			runCLIWithOutputs("set-flag-config",
				"--token", token,
				"--org-id", orgID,
				"--application-name", appName,
				"--flag-name", flagName,
				"--environment-name", envName,
				"--config", disableConfig)
		}

		// Then delete the flag
		cleanupOutput, cleanupOutputDir, cleanupErr := runCLIWithOutputs("delete-flag",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--confirm",
			"--verbose")

		defer os.RemoveAll(cleanupOutputDir)

		if cleanupErr != nil {
			t.Logf("Cleanup failed for flag %s: %v\nOutput: %s", flagName, cleanupErr, cleanupOutput)
		} else {
			t.Logf("Cleaned up test flag: %s", flagName)
		}
	})

	return flagName
}

// TestE2ECreateFlag tests the create-flag command following action pattern
func TestE2ECreateFlag(t *testing.T) {
	if !hasRequiredEnvVars(t) {
		t.Skip("Skipping E2E test - required environment variables not set")
	}

	token := os.Getenv("CLOUDBEES_TOKEN")
	orgID := os.Getenv("CLOUDBEES_ORG_ID")
	appName := os.Getenv("TEST_APPLICATION_NAME")

	// Test Boolean flag creation (like fm-create-flag action)
	flagName := fmt.Sprintf("e2e-boolean-flag-%d", time.Now().Unix())

	output, outputDir, err := runCLIWithOutputs("create-flag",
		"--token", token,
		"--org-id", orgID,
		"--application-name", appName,
		"--flag-name", flagName,
		"--flag-type", "Boolean",
		"--description", "E2E test Boolean flag",
		"--verbose")

	defer os.RemoveAll(outputDir)

	require.NoError(t, err)
	assert.Contains(t, output, "Successfully created flag")

	// Verify outputs match action expectations
	assert.True(t, outputExists(outputDir, "flag-id"))
	assert.True(t, outputExists(outputDir, "flag-name"))
	assert.True(t, outputExists(outputDir, "flag-type"))

	flagID, err := readOutput(outputDir, "flag-id")
	require.NoError(t, err)
	require.NotEmpty(t, flagID)

	createdFlagName, err := readOutput(outputDir, "flag-name")
	require.NoError(t, err)
	assert.Equal(t, flagName, createdFlagName)

	flagType, err := readOutput(outputDir, "flag-type")
	require.NoError(t, err)
	assert.Equal(t, "Boolean", flagType)

	// Cleanup
	t.Cleanup(func() {
		runCLI("delete-flag",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--confirm")
	})
}

// TestE2ECreateStringFlag tests creating a string flag with YAML variants
func TestE2ECreateStringFlag(t *testing.T) {
	if !hasRequiredEnvVars(t) {
		t.Skip("Skipping E2E test - required environment variables not set")
	}

	token := os.Getenv("CLOUDBEES_TOKEN")
	orgID := os.Getenv("CLOUDBEES_ORG_ID")
	appName := os.Getenv("TEST_APPLICATION_NAME")

	flagName := fmt.Sprintf("e2e-string-flag-%d", time.Now().Unix())

	// Test with YAML variants (like action would send)
	yamlVariants := `["option1", "option2", "option3"]`

	output, outputDir, err := runCLIWithOutputs("create-flag",
		"--token", token,
		"--org-id", orgID,
		"--application-name", appName,
		"--flag-name", flagName,
		"--flag-type", "String",
		"--description", "E2E test String flag with YAML variants",
		"--variants", yamlVariants,
		"--verbose")

	defer os.RemoveAll(outputDir)

	require.NoError(t, err)
	assert.Contains(t, output, "Successfully created flag")

	// Cleanup
	t.Cleanup(func() {
		runCLI("delete-flag",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--confirm")
	})
}

// TestE2EGetFlagConfig tests the get-flag-config command following action pattern
func TestE2EGetFlagConfig(t *testing.T) {
	if !hasRequiredEnvVars(t) {
		t.Skip("Skipping E2E test - required environment variables not set")
	}

	token := os.Getenv("CLOUDBEES_TOKEN")
	orgID := os.Getenv("CLOUDBEES_ORG_ID")
	appName := os.Getenv("TEST_APPLICATION_NAME")
	envName := os.Getenv("TEST_ENVIRONMENT_NAME")

	// Create a test flag first
	flagName := createTestFlag(t, token, orgID, appName)

	// Test getting flag config (like fm-get-flag-config action)
	_, outputDir, err := runCLIWithOutputs("get-flag-config",
		"--token", token,
		"--org-id", orgID,
		"--application-name", appName,
		"--flag-name", flagName,
		"--environment-name", envName,
		"--verbose")

	defer os.RemoveAll(outputDir)

	require.NoError(t, err)

	// Verify outputs match action expectations
	assert.True(t, outputExists(outputDir, "flag-config"))
	assert.True(t, outputExists(outputDir, "flag-id"))
	assert.True(t, outputExists(outputDir, "environment-id"))
	assert.True(t, outputExists(outputDir, "enabled"))
	assert.True(t, outputExists(outputDir, "default-value"))

	flagConfig, err := readOutput(outputDir, "flag-config")
	require.NoError(t, err)
	assert.Contains(t, flagConfig, "configuration")

	enabled, err := readOutput(outputDir, "enabled")
	require.NoError(t, err)
	assert.Contains(t, enabled, "false") // New flags default to disabled
}

// TestE2ESetFlagConfig tests the set-flag-config command with YAML config
func TestE2ESetFlagConfig(t *testing.T) {
	if !hasRequiredEnvVars(t) {
		t.Skip("Skipping E2E test - required environment variables not set")
	}

	token := os.Getenv("CLOUDBEES_TOKEN")
	orgID := os.Getenv("CLOUDBEES_ORG_ID")
	appName := os.Getenv("TEST_APPLICATION_NAME")
	envName := os.Getenv("TEST_ENVIRONMENT_NAME")

	// Create a test flag first
	flagName := createTestFlag(t, token, orgID, appName)

	// Test setting flag config with YAML (like fm-update-flag action)
	yamlConfig := `enabled: true
defaultValue: true`

	output, outputDir, err := runCLIWithOutputs("set-flag-config",
		"--token", token,
		"--org-id", orgID,
		"--application-name", appName,
		"--flag-name", flagName,
		"--environment-name", envName,
		"--config", yamlConfig,
		"--verbose")

	defer os.RemoveAll(outputDir)

	require.NoError(t, err)
	assert.Contains(t, output, "Successfully updated flag")

	// Verify outputs exist (matching action expectations)
	assert.True(t, outputExists(outputDir, "flag-id"))
	assert.True(t, outputExists(outputDir, "environment-id"))
	assert.True(t, outputExists(outputDir, "configuration"))
	assert.True(t, outputExists(outputDir, "success"))

	success, err := readOutput(outputDir, "success")
	require.NoError(t, err)
	assert.Equal(t, "true", success)
}

// TestE2EFullCRUDWorkflow tests the complete CRUD workflow that actions would perform
func TestE2EFullCRUDWorkflow(t *testing.T) {
	if !hasRequiredEnvVars(t) {
		t.Skip("Skipping E2E test - required environment variables not set")
	}

	token := os.Getenv("CLOUDBEES_TOKEN")
	orgID := os.Getenv("CLOUDBEES_ORG_ID")
	appName := os.Getenv("TEST_APPLICATION_NAME")
	envName := os.Getenv("TEST_ENVIRONMENT_NAME")

	flagName := fmt.Sprintf("e2e-crud-workflow-%d", time.Now().Unix())

	// Step 1: CREATE - Create a new flag (fm-create-flag action)
	t.Run("Create Flag", func(t *testing.T) {
		output, outputDir, err := runCLIWithOutputs("create-flag",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--flag-type", "Boolean",
			"--description", "E2E CRUD workflow test flag",
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)
		assert.Contains(t, output, "Successfully created flag")

		// Verify all action outputs
		assert.True(t, outputExists(outputDir, "flag-id"))
		assert.True(t, outputExists(outputDir, "flag-name"))
		assert.True(t, outputExists(outputDir, "flag-type"))
	})

	// Step 2: READ - Get initial flag configuration (fm-get-flag-config action)
	var initialConfig string
	t.Run("Get Initial Config", func(t *testing.T) {
		_, outputDir, err := runCLIWithOutputs("get-flag-config",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--environment-name", envName,
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)

		// Verify all action outputs
		assert.True(t, outputExists(outputDir, "flag-config"))
		assert.True(t, outputExists(outputDir, "flag-id"))
		assert.True(t, outputExists(outputDir, "environment-id"))
		assert.True(t, outputExists(outputDir, "enabled"))
		assert.True(t, outputExists(outputDir, "default-value"))

		initialConfig, err = readOutput(outputDir, "flag-config")
		require.NoError(t, err)
		assert.NotEmpty(t, initialConfig)

		enabled, err := readOutput(outputDir, "enabled")
		require.NoError(t, err)
		assert.Equal(t, "false", enabled) // Should start disabled
	})

	// Step 3: UPDATE - Enable the flag (fm-update-flag action)
	t.Run("Update Flag - Enable", func(t *testing.T) {
		yamlConfig := `enabled: true
defaultValue: true`

		output, outputDir, err := runCLIWithOutputs("set-flag-config",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--environment-name", envName,
			"--config", yamlConfig,
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)
		assert.Contains(t, output, "Successfully updated flag")

		// Verify action outputs
		assert.True(t, outputExists(outputDir, "success"))
		success, err := readOutput(outputDir, "success")
		require.NoError(t, err)
		assert.Equal(t, "true", success)
	})

	// Step 4: READ - Verify the update (fm-get-flag-config action)
	t.Run("Verify Update", func(t *testing.T) {
		_, outputDir, err := runCLIWithOutputs("get-flag-config",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--environment-name", envName,
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)

		enabled, err := readOutput(outputDir, "enabled")
		require.NoError(t, err)
		assert.Equal(t, "true", enabled) // Should now be enabled

		flagConfig, err := readOutput(outputDir, "flag-config")
		require.NoError(t, err)
		assert.NotEqual(t, initialConfig, flagConfig) // Config should have changed
	})

	// Step 5: UPDATE - Complex configuration with A/B testing
	t.Run("Update Flag - A/B Testing", func(t *testing.T) {
		yamlConfig := `enabled: true
defaultValue:
  - option: true
    percentage: 75
  - option: false
    percentage: 25`

		output, outputDir, err := runCLIWithOutputs("set-flag-config",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--environment-name", envName,
			"--config", yamlConfig,
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)
		assert.Contains(t, output, "Successfully updated flag")
	})

	// Step 6: READ - Verify A/B testing config
	t.Run("Verify A/B Config", func(t *testing.T) {
		_, outputDir, err := runCLIWithOutputs("get-flag-config",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--environment-name", envName,
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)

		flagConfig, err := readOutput(outputDir, "flag-config")
		require.NoError(t, err)
		assert.Contains(t, flagConfig, "percentage") // Should contain A/B testing config
	})

	// Step 7: DISABLE - Disable the flag before deletion (enabled flags can't be deleted)
	t.Run("Disable Flag", func(t *testing.T) {
		yamlConfig := `enabled: false
defaultValue: false`

		output, outputDir, err := runCLIWithOutputs("set-flag-config",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--environment-name", envName,
			"--config", yamlConfig,
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)
		assert.Contains(t, output, "Successfully updated flag")
	})

	// Step 8: DELETE - Clean up the flag (now that it's disabled)
	t.Run("Delete Flag", func(t *testing.T) {
		output, outputDir, err := runCLIWithOutputs("delete-flag",
			"--token", token,
			"--org-id", orgID,
			"--application-name", appName,
			"--flag-name", flagName,
			"--confirm",
			"--verbose")

		defer os.RemoveAll(outputDir)

		require.NoError(t, err)
		assert.Contains(t, output, "Successfully deleted flag")
	})
}

// TestCloudBeesOutputEnvironment tests that CLOUDBEES_OUTPUTS environment variable is handled
func TestCloudBeesOutputEnvironment(t *testing.T) {
	// Create temporary directory
	outputDir, err := ioutil.TempDir("", "cloudbees_outputs_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// Set CLOUDBEES_OUTPUTS environment variable
	cmd := exec.Command("./fm-actions", "list-environments", "--help")
	env := os.Environ()
	env = append(env, fmt.Sprintf("CLOUDBEES_OUTPUTS=%s", outputDir))
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Contains(t, string(output), "List all environments")
}

// TestInvalidAPI tests behavior with invalid API parameters
func TestInvalidAPI(t *testing.T) {
	output, outputDir, err := runCLIWithOutputs("list-environments",
		"--token=invalid-token",
		"--org-id=invalid-org-id")

	// Should fail with API error
	require.Error(t, err)
	assert.Contains(t, output, "failed")

	// Cleanup
	defer os.RemoveAll(outputDir)
}

// TestDryRunFunctionality tests dry-run across commands
func TestDryRunFunctionality(t *testing.T) {
	// Test create-flag dry-run
	t.Run("create-flag dry-run", func(t *testing.T) {
		output, err := runCLI("create-flag",
			"--token=test-token",
			"--org-id=test-org",
			"--application-name=test-app",
			"--flag-name=test-flag",
			"--dry-run")

		require.NoError(t, err)
		assert.Contains(t, output, "DRY RUN:")
	})

	// Test set-flag-config dry-run
	t.Run("set-flag-config dry-run", func(t *testing.T) {
		output, err := runCLI("set-flag-config",
			"--token=test-token",
			"--org-id=test-org",
			"--application-name=test-app",
			"--flag-name=test-flag",
			"--environment-name=test-env",
			"--enabled=true",
			"--dry-run")

		require.NoError(t, err)
		assert.Contains(t, output, "DRY RUN:")
	})
}
