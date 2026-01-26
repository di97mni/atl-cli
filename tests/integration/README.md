# Integration Tests

Integration tests for atl-cli require real Atlassian Cloud credentials.

## Setup

1. Set environment variables:
   ```bash
   export ATL_CLI_SITE="your-site.atlassian.net"
   export ATL_CLI_EMAIL="your-email@example.com"
   export ATL_CLI_TOKEN="your-api-token"
   ```

2. Generate an API token at: https://id.atlassian.com/manage-profile/security/api-tokens

## Running Tests

```bash
go test ./tests/integration/...
```

## Test Data Requirements

- Access to at least one Jira project with existing issues
- Access to at least one Confluence space with existing pages
- Read permissions for test resources

## CI/CD

Integration tests are not run in CI by default. They require:
- Secret environment variables configured
- Network access to *.atlassian.net

To run in CI, set the `RUN_INTEGRATION_TESTS=true` environment variable.
