# CloudBees Feature Management Actions Container

This container provides the underlying implementation for CloudBees Feature Management actions. It contains various commands that can be used to interact with CloudBees Platform Feature Management.

## CloudBees Actions Available

This workspace provides the following CloudBees Actions for Feature Management:

- **[fm-create-flag](https://github.com/cloudbees-days/fm-create-flag)** - Create a new feature flag in an application
- **[fm-get-flag-config](https://github.com/cloudbees-days/fm-get-flag-config)** - Get current feature flag configuration from an environment  
- **[fm-update-flag](https://github.com/cloudbees-days/fm-update-flag)** - Update feature flag configuration in an environment

## Container Commands

The container includes several commands that power the CloudBees Actions above:

- `create-flag` - Used by fm-create-flag action
- `get-flag-config` - Used by fm-get-flag-config action  
- `set-flag-config` - Used by fm-update-flag action
- `list-environments` - Helper command for listing environments
- `list-flags` - Helper command for listing flags
- `delete-flag` - Helper command for deleting flags

## Setup Requirements

All actions require these CloudBees Platform connection details:

- `token` - Your CloudBees Platform API token (Personal Access Token)
- `org-id` - Your organization ID (UUID)  
- `api-url` - CloudBees Platform API URL (defaults to `https://api.cloudbees.io`)

### Getting a CloudBees Platform API Token

1. Go to your CloudBees Platform user profile
2. Navigate to the security/API tokens section
3. Click "Create API token" 
4. Use this token as the `token` input

## Development

This container is built as a Docker image and used by the CloudBees Actions above. Each action calls specific commands within this container to perform Feature Management operations.
