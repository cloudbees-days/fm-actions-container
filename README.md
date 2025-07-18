# CloudBees Feature Management Actions

A collection of CloudBees actions for interacting with CloudBees Platform Feature Management in your workflows.

## Available Actions

- **`list-environments`** - List all environments in your organization
- **`list-flags`** - List all feature flags in an application
- **`create-flag`** - Create a new feature flag
- **`get-flag-config`** - Get current feature flag configuration for a specific environment
- **`set-flag-config`** - Set feature flag configuration (enable/disable, set values) for a target environment
- **`delete-flag`** - Delete an existing feature flag

## Setup

### Required Inputs

All actions require these CloudBees Platform connection details:

- `token` - Your CloudBees Platform API token (Personal Access Token)
- `org-id` - Your organization ID (UUID)
- `api-url` - CloudBees Platform API URL (defaults to `https://api.cloudbees.io`)

### Getting a CloudBees Platform API Token

1. Go to your CloudBees Platform user profile
2. Navigate to the security/API tokens section
3. Click "Create API token"
4. Use this token as the `token` input

## Commands

### list-environments

List all environments in your organization.

```bash
./fm-actions list-environments \
  --token="your-api-token" \
  --org-id="your-org-id"
```

**Outputs:**
- `environment-count` - Number of environments found
- `environments` - JSON array of environment objects

### list-flags

List all feature flags in an application.

```bash
./fm-actions list-flags \
  --token="your-api-token" \
  --org-id="your-org-id" \
  --application-name="my-app"
```

**Outputs:**
- `flag-count` - Number of flags found
- `flags` - JSON array of flag objects

### create-flag

Create a new feature flag in an application.

```bash
./fm-actions create-flag \
  --token="your-api-token" \
  --org-id="your-org-id" \
  --application-name="my-app" \
  --flag-name="new-feature" \
  --flag-type="Boolean" \
  --description="New feature toggle"
```

**Options:**
- `--flag-type` - Type of flag: Boolean, String, Number (default: Boolean)
- `--description` - Description of the flag
- `--variants` - Variants as YAML array or comma-separated list
- `--is-permanent` - Whether the flag is permanent
- `--dry-run` - Validate flag details without creating

**Outputs:**
- `flag-id` - The created flag's unique ID
- `flag-name` - The flag name
- `flag-type` - The flag type
- `flag` - Full flag details as JSON
- `success` - Whether the operation succeeded

### get-flag-config

Get current feature flag configuration for a specific flag in an environment.

```bash
./fm-actions get-flag-config \
  --token="your-api-token" \
  --org-id="your-org-id" \
  --application-name="my-app" \
  --flag-name="feature-flag-name" \
  --environment-name="production"
```

**Outputs:**
- `flag-config` - The current flag configuration as JSON
- `flag-id` - The flag's unique ID
- `environment-id` - The environment's unique ID
- `enabled` - Whether the flag is enabled (true/false)
- `default-value` - The default value of the flag

### set-flag-config

Set feature flag configuration for a target environment.

```bash
./fm-actions set-flag-config \
  --token="your-api-token" \
  --org-id="your-org-id" \
  --application-name="my-app" \
  --flag-name="feature-flag-name" \
  --environment-name="production" \
  --enabled="true" \
  --default-value="true"
```

**Options:**
- `--enabled` - Enable/disable the flag (true/false)
- `--default-value` - Default value for the flag (JSON or string)
- `--variants-enabled` - Enable/disable variants (true/false)
- `--stickiness-property` - Stickiness property for consistent evaluation
- `--config` - Complete configuration as YAML
- `--dry-run` - Validate configuration without applying changes

**Outputs:**
- `flag-id` - The flag's unique ID
- `flag-name` - The flag name
- `application-id` - The application's unique ID
- `application-name` - The application name
- `environment-id` - The environment's unique ID
- `environment-name` - The environment name
- `configuration` - Updated configuration as JSON
- `enabled` - Whether the flag is enabled (true/false)
- `success` - Whether the operation succeeded (true/false)

### delete-flag

Delete an existing feature flag from an application.

```bash
./fm-actions delete-flag \
  --token="your-api-token" \
  --org-id="your-org-id" \
  --application-name="my-app" \
  --flag-name="old-feature" \
  --confirm
```

**Options:**
- `--dry-run` - Preview the deletion without actually deleting
- `--confirm` - Confirm that you want to delete the flag (required unless using dry-run)

**Note:** Enabled flags cannot be deleted. You must first disable the flag in all environments before deletion.

**Outputs:**
- `flag-id` - The deleted flag's unique ID
- `flag-name` - The flag name
- `success` - Whether the operation succeeded
