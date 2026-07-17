# Command Reference Guide

This file is generated from live CLI help output.
For authoritative command behavior, also use:

```bash
gpd --help
gpd <command> --help
gpd <command> <subcommand> --help
```

To regenerate:

```bash
make generate-command-docs
```

Generated: 2026-07-17T05:31:40Z

## Global help

```
Usage: gpd <command> [flags]

Google Play Developer CLI - A fast, lightweight command-line interface for the
Google Play Developer Console.

Flags:
  -h, --help                   Show context-sensitive help.
  -p, --package=STRING         App package name
      --output="json"          Output format: json, table, markdown, csv,
                               excel (default: table on TTY, json in pipes/CI;
                               override with GPD_DEFAULT_OUTPUT)
      --pretty                 Pretty print JSON output
      --timeout=30s            Network timeout
      --store-tokens="auto"    Token storage: auto, never, secure
      --fields=STRING          JSON field projection (comma-separated paths)
      --quiet                  Suppress non-error output
  -v, --verbose                Enable verbose logging
      --key-path=STRING        Path to service account key file
      --profile=STRING         Configuration profile to use
      --cache-dir=STRING       Cache directory for temporary data
                               ($GPD_CACHE_DIR)

Commands:
  auth status                      Check authentication status
  auth login                       Authenticate with Google Play
  auth init                        Initialize auth for a profile (alias of
                                   login)
  auth logout                      Sign out and clear stored credentials for a
                                   profile
  auth delete                      Delete a stored authentication profile
  auth list                        List stored authentication profiles
  auth switch                      Switch the active authentication profile
  auth check                       Validate package permissions for current
                                   credentials
  auth doctor                      Diagnose authentication setup
  auth diagnose                    Detailed auth diagnostics (alias of doctor)
  config init                      Initialize project configuration
  config doctor                    Diagnose configuration and credential issues
  config path                      Show configuration paths
  config get                       Get configuration value
  config set                       Set configuration value
  config print                     Print current configuration
  config export                    Export configuration to file
  config import                    Import configuration from file
  config completion                Generate shell completion script
  publish play                     High-level upload→track→status publish job
                                   (ASC publish analogue)
  publish upload                   Upload APK or AAB
  publish release                  Create or update a release
  publish rollout                  Update rollout percentage
  publish promote                  Promote a release between tracks
  publish halt                     Halt a production rollout
  publish rollback                 Rollback to a previous version
  publish status                   Get track status
  publish tracks                   List all tracks
  publish capabilities             List publishing capabilities
  publish listing update           Update store listing
  publish listing get              Get store listing
  publish listing delete           Delete store listing
  publish details get              Get app details
  publish details update           Update app details
  publish details patch            Patch app details
  publish images upload            Upload an image
  publish images list              List images
  publish images delete            Delete an image
  publish images delete-all        Delete all images for type
  publish assets upload            Upload assets from directory
  publish assets spec              Output asset validation matrix
  publish deobfuscation upload     Upload deobfuscation file
  publish testers add              Add tester groups
  publish testers remove           Remove tester groups
  publish testers list             List tester groups
  publish testers get              Get tester groups for a track
  publish builds list              List uploaded builds
  publish builds get               Get build details
  publish builds expire            Expire a build from tracks
  publish builds expire-all        Expire all builds from tracks
  publish beta-groups list         List beta groups
  publish beta-groups get          Get beta group details
  publish beta-groups create       Create beta group
  publish beta-groups update       Update beta group testers
  publish beta-groups delete       Delete beta group
  publish beta-groups add-testers
                                   Add tester Google Groups to a beta group
  publish beta-groups remove-testers
                                   Remove tester Google Groups from a beta group
  publish internal-share upload    Upload artifact for internal sharing
  reviews list                     List user reviews
  reviews get                      Get a review by ID
  reviews reply                    Reply to a review
  reviews response-get             Get response for a review
  reviews response-delete          Delete response for a review
  vitals crashes                   Query crash rate data
  vitals anrs                      Query ANR rate data
  vitals errors issues             Search error issues
  vitals errors reports            Search error reports
  vitals errors counts get         Get error count metrics
  vitals errors counts query       Query error counts over time
  vitals metrics excessive-wakeups
                                   Query excessive wakeups data
  vitals metrics slow-rendering    Query slow rendering data
  vitals metrics slow-start        Query slow start data
  vitals metrics stuck-wakelocks
                                   Query stuck wakelocks data
  vitals anomalies list            List anomalies
  vitals query                     Query vitals metrics
  vitals capabilities              List available vitals metrics
  monitor watch                    Continuous vitals monitoring with threshold
                                   alerts
  monitor anomalies                Detect statistical anomalies in vitals
                                   metrics
  monitor dashboard                Generate monitoring dashboard data
  monitor report                   Generate scheduled monitoring reports
  monitor webhooks list            List configured webhooks (simulated)
  analytics query                  Query analytics data
  analytics capabilities           List analytics capabilities
  purchases products acknowledge
                                   Acknowledge a product purchase
  purchases products consume       Consume a product purchase
  purchases subscriptions acknowledge
                                   Acknowledge a subscription purchase
  purchases subscriptions cancel
                                   Cancel a subscription
  purchases subscriptions defer    Defer a subscription renewal
  purchases subscriptions refund
                                   Refund a subscription
  purchases subscriptions revoke
                                   Revoke a subscription
  purchases verify                 Verify purchase
  purchases voided list            List voided purchases
  purchases capabilities           List purchase verification capabilities
  monetization products list       List in-app products
  monetization products get        Get an in-app product
  monetization products create     Create an in-app product
  monetization products update     Update an in-app product
  monetization products delete     Delete an in-app product
  monetization subscriptions list
                                   List subscription products
  monetization subscriptions get
                                   Get a subscription product
  monetization subscriptions create
                                   Create a subscription
  monetization subscriptions update
                                   Update a subscription
  monetization subscriptions patch
                                   Patch a subscription
  monetization subscriptions delete
                                   Delete a subscription
  monetization subscriptions archive
                                   Archive a subscription
  monetization subscriptions batch-get
                                   Batch get subscriptions
  monetization subscriptions batch-update
                                   Batch update subscriptions
  monetization one-time-products list
                                   List one-time products
  monetization one-time-products get
                                   Get a one-time product
  monetization one-time-products create
                                   Create a one-time product
  monetization one-time-products update
                                   Update a one-time product
  monetization one-time-products delete
                                   Delete a one-time product
  monetization one-time-products batch-get
                                   Batch get one-time products
  monetization one-time-products batch-update
                                   Batch update one-time products
  monetization base-plans activate
                                   Activate a base plan
  monetization base-plans deactivate
                                   Deactivate a base plan
  monetization base-plans delete
                                   Delete a base plan
  monetization base-plans migrate-prices
                                   Migrate base plan prices
  monetization base-plans batch-migrate
                                   Batch migrate base plan prices
  monetization base-plans batch-update-states
                                   Batch update base plan states
  monetization offers create       Create an offer
  monetization offers get          Get an offer
  monetization offers list         List offers
  monetization offers delete       Delete an offer
  monetization offers activate     Activate an offer
  monetization offers deactivate
                                   Deactivate an offer
  monetization offers batch-get    Batch get offers
  monetization offers batch-update
                                   Batch update offers
  monetization offers batch-update-states
                                   Batch update offer states
  monetization capabilities        List monetization capabilities
  permissions users add            Add a user
  permissions users remove         Remove a user
  permissions users list           List users
  permissions grants add           Add a grant
  permissions grants remove        Remove a grant
  permissions grants list          List grants
  permissions list                 List permissions
  recovery list                    List recovery actions
  recovery create                  Create recovery action
  recovery deploy                  Deploy recovery
  recovery cancel                  Cancel recovery
  apps list                        List apps in the developer account
  apps get                         Get app details
  games achievements reset         Reset achievements
  games scores reset               Reset scores on a leaderboard
  games events reset               Reset game events
  games players hide               Hide a player
  games players unhide             Unhide a player
  games capabilities               List Games management capabilities
  integrity decode                 Decode integrity token
  migrate                          Migration commands
  custom-app (customapp)           Custom app publishing
  generated-apks list              List generated APK variants for a bundle
  generated-apks download          Download a generated APK
  system-apks variants list        List system APK variants for a version code
  system-apks variants get         Get a specific system APK variant
  system-apks variants create      Create a system APK variant for a device spec
  system-apks variants download    Download a system APK variant
  grouping                         App access grouping
  version                          Show version information
  check-update                     Check for available updates
  completion                       Generate shell completion scripts
  maintenance drift                Detect API drift between discovery and client
                                   library
  maintenance multi-drift          Monitor drift across multiple Google APIs
  maintenance health               Check system health and dependencies
  maintenance update-check         Check for CLI updates
  bulk upload                      Upload multiple APKs/AABs in parallel
  bulk listings                    Update store listings across multiple locales
  bulk images                      Batch upload images for multiple types
  bulk tracks                      Update multiple tracks at once
  compare vitals                   Compare vitals metrics across multiple apps
  compare reviews                  Compare review metrics across apps
  compare releases                 Compare release history across apps
  compare subscriptions            Compare subscription metrics
  release-mgmt calendar            Show upcoming and past releases
  release-mgmt conflicts           Detect version code conflicts
  release-mgmt strategy            Get rollback/roll-forward recommendations
  release-mgmt history             Show detailed release history
  release-mgmt notes               Manage release notes across locales
  testing prelaunch                Trigger or check pre-launch report
  testing device-lab               Run tests on Firebase Test Lab
  testing screenshots              Capture screenshots across devices
  testing validate                 Comprehensive app validation
  testing compatibility            Check device compatibility
  automation release-notes         Generate release notes from git history or
                                   PRs
  automation rollout               Automated staged rollout with health checks
  automation promote               Smart promote with optional verification
  automation validate              Comprehensive pre-release validation
  automation monitor               Monitor release health after rollout
  workflow run                     Execute a workflow from a JSON file
  workflow list                    List available workflows and run history
  workflow show                    Show workflow definition and details
  workflow status                  Show status of a workflow run
  workflow init                    Create a new workflow from template
  workflow logs                    Show logs from a workflow run step
  workflow validate                Validate workflow file for errors
  validate                         Submission readiness / pre-publish validation
                                   report
  extension install                Install an extension
  extension list                   List installed extensions
  extension remove                 Remove an extension
  extension upgrade                Upgrade an extension
  extension exec                   Execute an extension explicitly

Run "gpd <command> --help" for more information on a command.

Extension Commands:
  gpd test-ext  Extension command
```

## Auth

```
Usage: gpd auth <command> [flags]

Authentication commands

Flags:
  -h, --help                   Show context-sensitive help.
  -p, --package=STRING         App package name
      --output="json"          Output format: json, table, markdown, csv,
                               excel (default: table on TTY, json in pipes/CI;
                               override with GPD_DEFAULT_OUTPUT)
      --pretty                 Pretty print JSON output
      --timeout=30s            Network timeout
      --store-tokens="auto"    Token storage: auto, never, secure
      --fields=STRING          JSON field projection (comma-separated paths)
      --quiet                  Suppress non-error output
  -v, --verbose                Enable verbose logging
      --key-path=STRING        Path to service account key file
      --profile=STRING         Configuration profile to use
      --cache-dir=STRING       Cache directory for temporary data
                               ($GPD_CACHE_DIR)

Commands:
  auth status      Check authentication status
  auth login       Authenticate with Google Play
  auth init        Initialize auth for a profile (alias of login)
  auth logout      Sign out and clear stored credentials for a profile
  auth delete      Delete a stored authentication profile
  auth list        List stored authentication profiles
  auth switch      Switch the active authentication profile
  auth check       Validate package permissions for current credentials
  auth doctor      Diagnose authentication setup
  auth diagnose    Detailed auth diagnostics (alias of doctor)

Extension Commands:
  gpd test-ext  Extension command
```

## Publish

```
Usage: gpd publish <command> [flags]

Publishing commands

Flags:
  -h, --help                   Show context-sensitive help.
  -p, --package=STRING         App package name
      --output="json"          Output format: json, table, markdown, csv,
                               excel (default: table on TTY, json in pipes/CI;
                               override with GPD_DEFAULT_OUTPUT)
      --pretty                 Pretty print JSON output
      --timeout=30s            Network timeout
      --store-tokens="auto"    Token storage: auto, never, secure
      --fields=STRING          JSON field projection (comma-separated paths)
      --quiet                  Suppress non-error output
  -v, --verbose                Enable verbose logging
      --key-path=STRING        Path to service account key file
      --profile=STRING         Configuration profile to use
      --cache-dir=STRING       Cache directory for temporary data
                               ($GPD_CACHE_DIR)

Commands:
  publish play                     High-level upload→track→status publish job
                                   (ASC publish analogue)
  publish upload                   Upload APK or AAB
  publish release                  Create or update a release
  publish rollout                  Update rollout percentage
  publish promote                  Promote a release between tracks
  publish halt                     Halt a production rollout
  publish rollback                 Rollback to a previous version
  publish status                   Get track status
  publish tracks                   List all tracks
  publish capabilities             List publishing capabilities
  publish listing update           Update store listing
  publish listing get              Get store listing
  publish listing delete           Delete store listing
  publish details get              Get app details
  publish details update           Update app details
  publish details patch            Patch app details
  publish images upload            Upload an image
  publish images list              List images
  publish images delete            Delete an image
  publish images delete-all        Delete all images for type
  publish assets upload            Upload assets from directory
  publish assets spec              Output asset validation matrix
  publish deobfuscation upload     Upload deobfuscation file
  publish testers add              Add tester groups
  publish testers remove           Remove tester groups
  publish testers list             List tester groups
  publish testers get              Get tester groups for a track
  publish builds list              List uploaded builds
  publish builds get               Get build details
  publish builds expire            Expire a build from tracks
  publish builds expire-all        Expire all builds from tracks
  publish beta-groups list         List beta groups
  publish beta-groups get          Get beta group details
  publish beta-groups create       Create beta group
  publish beta-groups update       Update beta group testers
  publish beta-groups delete       Delete beta group
  publish beta-groups add-testers
                                   Add tester Google Groups to a beta group
  publish beta-groups remove-testers
                                   Remove tester Google Groups from a beta group
  publish internal-share upload    Upload artifact for internal sharing

Extension Commands:
  gpd test-ext  Extension command
```

## Validate

```
Usage: gpd validate [flags]

Submission readiness / pre-publish validation report

Flags:
  -h, --help                   Show context-sensitive help.
  -p, --package=STRING         App package name
      --output="json"          Output format: json, table, markdown, csv,
                               excel (default: table on TTY, json in pipes/CI;
                               override with GPD_DEFAULT_OUTPUT)
      --pretty                 Pretty print JSON output
      --timeout=30s            Network timeout
      --store-tokens="auto"    Token storage: auto, never, secure
      --fields=STRING          JSON field projection (comma-separated paths)
      --quiet                  Suppress non-error output
  -v, --verbose                Enable verbose logging
      --key-path=STRING        Path to service account key file
      --profile=STRING         Configuration profile to use
      --cache-dir=STRING       Cache directory for temporary data
                               ($GPD_CACHE_DIR)

      --track="internal"       Target track for the readiness plan
      --file=STRING            Optional APK/AAB path to validate locally
      --strict                 Treat warnings as failures
      --dry-run                Plan checks without network side effects (default
                               true)
      --network                Opt-in network probes: package access,
                               track list, listing (requires --package and
                               credentials)

Extension Commands:
  gpd test-ext  Extension command
```

## Workflow

```
Usage: gpd workflow <command> [flags]

Declarative workflow execution

Flags:
  -h, --help                   Show context-sensitive help.
  -p, --package=STRING         App package name
      --output="json"          Output format: json, table, markdown, csv,
                               excel (default: table on TTY, json in pipes/CI;
                               override with GPD_DEFAULT_OUTPUT)
      --pretty                 Pretty print JSON output
      --timeout=30s            Network timeout
      --store-tokens="auto"    Token storage: auto, never, secure
      --fields=STRING          JSON field projection (comma-separated paths)
      --quiet                  Suppress non-error output
  -v, --verbose                Enable verbose logging
      --key-path=STRING        Path to service account key file
      --profile=STRING         Configuration profile to use
      --cache-dir=STRING       Cache directory for temporary data
                               ($GPD_CACHE_DIR)

Commands:
  workflow run         Execute a workflow from a JSON file
  workflow list        List available workflows and run history
  workflow show        Show workflow definition and details
  workflow status      Show status of a workflow run
  workflow init        Create a new workflow from template
  workflow logs        Show logs from a workflow run step
  workflow validate    Validate workflow file for errors

Extension Commands:
  gpd test-ext  Extension command
```

## Command families (top-level)

- `auth`
- `auth`
- `auth`
- `auth`
- `auth`
- `auth`
- `auth`
- `auth`
- `auth`
- `auth`
- `config`
- `config`
- `config`
- `config`
- `config`
- `config`
- `config`
- `config`
- `config`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `publish`
- `reviews`
- `reviews`
- `reviews`
- `reviews`
- `reviews`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `vitals`
- `monitor`
- `monitor`
- `monitor`
- `monitor`
- `monitor`
- `analytics`
- `analytics`
- `purchases`
- `purchases`
- `purchases`
- `purchases`
- `purchases`
- `purchases`
- `purchases`
- `purchases`
- `purchases`
- `purchases`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `monetization`
- `permissions`
- `permissions`
- `permissions`
- `permissions`
- `permissions`
- `permissions`
- `permissions`
- `recovery`
- `recovery`
- `recovery`
- `recovery`
- `apps`
- `apps`
- `games`
- `games`
- `games`
- `games`
- `games`
- `games`
- `integrity`
- `migrate`
- `custom-app`
- `generated-apks`
- `generated-apks`
- `system-apks`
- `system-apks`
- `system-apks`
- `system-apks`
- `grouping`
- `version`
- `check-update`
- `completion`
- `maintenance`
- `maintenance`
- `maintenance`
- `maintenance`
- `bulk`
- `bulk`
- `bulk`
- `bulk`
- `compare`
- `compare`
- `compare`
- `compare`
- `release-mgmt`
- `release-mgmt`
- `release-mgmt`
- `release-mgmt`
- `release-mgmt`
- `testing`
- `testing`
- `testing`
- `testing`
- `testing`
- `automation`
- `automation`
- `automation`
- `automation`
- `automation`
- `workflow`
- `workflow`
- `workflow`
- `workflow`
- `workflow`
- `workflow`
- `workflow`
- `validate`
- `extension`
- `extension`
- `extension`
- `extension`
- `extension`

