# Documentation Index

Welcome to the Google Play Developer CLI documentation. This directory contains guides, examples, and reference materials to help you get the most out of `gpd`.

## Overview

The Google Play Developer CLI (`gpd`) is a fast, lightweight command-line interface for the Google Play Developer Console. This documentation covers everything from getting started to advanced workflows and troubleshooting.

## Getting Started

### Main Documentation

- **[Main README](../README.md)** - Installation, quick start, and command reference

### Authentication Setup

Before using `gpd`, you'll need to set up authentication with a Google Cloud service account:

1. Create a service account in Google Cloud Console
2. Enable the Google Play Android Publisher API
3. Add the service account to your Play Console with appropriate permissions
4. Configure credentials using one of these methods:

```bash
# Option 1: Environment variable
export GPD_SERVICE_ACCOUNT_KEY='{"type": "service_account", ...}'

# Option 2: Key file
gpd --key /path/to/service-account.json auth status

# Option 3: Application Default Credentials
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
gpd auth status
```

### First Commands

Once authenticated, verify your setup:

```bash
# Check authentication status
gpd auth status

# Check permissions for a specific app
gpd auth check --package com.example.app

# Diagnose configuration issues
gpd config doctor
```

Then try your first commands:

```bash
# Upload an app bundle
gpd publish upload app.aab --package com.example.app

# List reviews
gpd reviews list --package com.example.app --min-rating 1 --max-rating 3
```

## Guides

### [Edit Workflow](examples/edit-workflow.md)

Learn how to manage edits for atomic releases. Edits are transactional units that allow you to make multiple changes before committing them, providing atomicity, validation, and rollback capabilities.

**Topics covered:**
- Understanding edit transactions
- Creating and managing edits
- Uploading artifacts within edits
- Validating and committing changes
- Multi-step release workflows

### [Subscription Management](examples/subscription-management.md)

Comprehensive guide to monetization features, including subscription management, base plans, offers, and regional pricing.

**Topics covered:**
- Creating and managing subscriptions
- Base plan configuration
- Offer management
- Batch operations
- Regional pricing conversion

### [CI/CD Integration](examples/ci-cd-integration.md)

Integrate `gpd` into your continuous integration and deployment pipelines with GitHub Actions, GitLab CI, and other CI/CD platforms.

**Topics covered:**
- GitHub Actions workflows
- GitLab CI configuration
- Secure credential management
- Automated release workflows
- Best practices for CI/CD

### [Error Debugging](examples/error-debugging.md)

Debug issues using Android Vitals and error reporting features. Learn how to query crashes, ANRs, and performance metrics.

**Topics covered:**
- Querying crash and ANR data
- Error search and reporting
- Performance metrics (excessive wakeups, slow rendering, slow start, stuck wakelocks)
- Anomalies detection
- Troubleshooting common issues

## Reference

### [API Coverage Matrix](api-coverage-matrix.md)

Complete reference of supported Google Play Developer API endpoints and their `gpd` command equivalents.

**Includes:**
- Publishing endpoints
- Review management
- Analytics and vitals
- Monetization APIs
- Purchase verification
- Permissions and access control

### [Migration Guides](migration/)

Upgrade guides for migrating between versions or adapting to API changes.

## Migration Guides

### [Assets to Images](migration/assets-to-images.md)

Guide for migrating from the legacy Assets API to the modern Images API for uploading app screenshots, icons, and graphics.

**Topics covered:**
- Differences between Assets and Images APIs
- Step-by-step migration process
- Command equivalents
- Common pitfalls and solutions

## Additional Resources

### GitHub Repository

- **Repository**: [github.com/dl-alexandre/gpd](https://github.com/dl-alexandre/gpd)
- **Issues**: [Report bugs or request features](https://github.com/dl-alexandre/gpd/issues)
- **Releases**: [Latest releases and changelog](https://github.com/dl-alexandre/gpd/releases)

### Contributing

- **[Contributing Guide](../CONTRIBUTING.md)** - Guidelines for contributing code, reporting bugs, and suggesting features

### Support

- **Documentation Issues**: Open an issue on GitHub with the `documentation` label
- **Questions**: Start a discussion in the GitHub Discussions section
- **Bug Reports**: Follow the guidelines in [CONTRIBUTING.md](../CONTRIBUTING.md#reporting-bugs)

---

**Quick Links:**
- [Command Reference](../README.md#command-reference)
- [Output Format](../README.md#output-format)
- [Configuration](../README.md#configuration)
- [AI Agent Integration](../README.md#ai-agent-integration)
