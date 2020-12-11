
> 🏗 This is a work-in-progress

# Orca

### Credential and secret scanning for Github Repositories

## Roadmap

- GitHub App boilerplate
- Automatically check on push, PR, or issue submitted
- Scan for:
    - Passwords
    - API keys / Personal Access Tokens
    - Connection strings
    - Generic credentials
    - Cryptocurrency wallet keys
    - Certificates / Private keys
    - Production configuration and environment files
        - Common files from popular frameworks (e.g. Laravel, Node, ASP.NET Core)
    - CI/CD process definition files
- Integrate with:
    - Password managers
        - Compare hashes?
    - CI/CD tools
        - Search for secrets defined in CI/CD configurations
- Automated response
    - Nuke key pairs
    - Squash commits on merge to mask secret in previous commit(s)?
    - Alert appropriate actors
- Independent CLI tool
- Git hooks

## Resources
- Setting up the Development Environment: https://docs.github.com/en/free-pro-team@latest/developers/apps/setting-up-your-development-environment-to-create-a-github-app
- GitHub WebHook Events and Payloads: https://docs.github.com/en/free-pro-team@latest/developers/webhooks-and-events/webhook-events-and-payloads
- GitHub REST API library (Octokit): https://developer.github.com/v3/libraries/ 
- Remediate sensitive data found in commits: https://stackoverflow.com/questions/872565/remove-sensitive-files-and-their-commits-from-git-history
- Standard Go project layout: https://github.com/golang-standards/project-layout
- CLI library: https://github.com/urfave/cli
- AWS Git Secrets: https://github.com/awslabs/git-secrets
