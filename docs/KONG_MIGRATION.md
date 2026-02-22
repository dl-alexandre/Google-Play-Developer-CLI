# Kong CLI Migration Guide

## Overview

Migration from Cobra to Kong CLI framework is in progress. Kong provides a more declarative approach using struct tags instead of programmatic command building.

## Current Status

✅ **Completed:**
- Kong dependency added to go.mod
- Base CLI structure (`kong_cli.go`) with global flags
- Foundation files for command types
- Helper utilities (`kong_helpers.go`)
- Placeholder command structs defined
- Main entry point updated to use Kong

⏳ **Pending:**
- Implement Run() methods for 41 remaining command groups
- Remove Cobra dependency

## Migration Pattern

### Before (Cobra):
```go
func (c *CLI) addAuthCommands() {
    authCmd := &cobra.Command{
        Use:   "auth",
        Short: "Authentication commands",
    }
    
    statusCmd := &cobra.Command{
        Use:   "status",
        Short: "Check authentication status",
        RunE: func(cmd *cobra.Command, args []string) error {
            return c.authStatus(cmd.Context())
        },
    }
    authCmd.AddCommand(statusCmd)
    c.rootCmd.AddCommand(authCmd)
}

func (c *CLI) authStatus(ctx context.Context) error {
    // Implementation
}
```

### After (Kong):
```go
// In kong_auth.go:
type AuthCmd struct {
    Status AuthStatusCmd `cmd:"" help:"Check authentication status"`
}

type AuthStatusCmd struct{}

func (cmd *AuthStatusCmd) Run(globals *Globals) error {
    ctx := context.Background()
    // Implementation using globals.Package, globals.Output, etc.
    return nil
}
```

## Command Structure

Commands are defined as nested structs with tags:
- `cmd:""` - Marks a struct as a command
- `help:"description"` - Help text for the command
- `arg:""` - Positional argument
- `optional:""` - Optional argument
- `type:"existingfile"` - File path validation
- `enum:"a,b,c"` - Allowed values
- `default:"value"` - Default value
- `short:"x"` - Short flag name

## Files to Create/Migrate

Create files in `internal/cli/` following the pattern `kong_<group>.go`:

1. `kong_config.go` - Config commands (init, doctor, get, set, etc.)
2. `kong_publish.go` - Publish commands (upload, release, etc.)
3. `kong_reviews.go` - Reviews commands
4. `kong_vitals.go` - Vitals commands
5. `kong_analytics.go` - Analytics commands
6. `kong_purchases.go` - Purchases commands
7. `kong_monetization.go` - Monetization commands
8. `kong_permissions.go` - Permissions commands
9. `kong_recovery.go` - Recovery commands
10. `kong_apps.go` - Apps commands
11. `kong_games.go` - Games commands
12. `kong_integrity.go` - Integrity commands
13. `kong_migrate.go` - Migrate commands
14. `kong_customapp.go` - CustomApp commands
15. `kong_grouping.go` - Grouping commands

## Implementation Steps

For each command group:

1. **Define the command struct:**
```go
type PublishCmd struct {
    Upload  PublishUploadCmd  `cmd:"" help:"Upload an artifact"`
    Release PublishReleaseCmd `cmd:"" help:"Create a release"`
}
```

2. **Define subcommand structs:**
```go
type PublishUploadCmd struct {
    File     string `arg:"" help:"File to upload" type:"existingfile"`
    Track    string `help:"Release track" default:"internal" enum:"internal,alpha,beta,production"`
    EditID   string `help:"Edit transaction ID"`
}
```

3. **Implement the Run() method:**
```go
func (cmd *PublishUploadCmd) Run(globals *Globals) error {
    // 1. Validate inputs
    if cmd.File == "" {
        return errors.NewAPIError(errors.CodeValidationError, "file is required")
    }
    
    // 2. Get context
    ctx := context.Background()
    
    // 3. Use globals for shared state
    packageName := globals.Package
    if packageName == "" {
        return errors.NewAPIError(errors.CodeValidationError, "package is required")
    }
    
    // 4. Call existing implementation
    // (You may need to refactor existing cobra command handlers)
    
    // 5. Output result
    result := output.NewResult(map[string]interface{}{
        "file": cmd.File,
        "track": cmd.Track,
    })
    
    return outputResult(result, globals.Output, globals.Pretty)
}
```

## Helper Functions

Available in `kong_helpers.go`:
- `newAuthManager()` - Create auth manager
- `newConfig()` - Load config
- `outputResult()` - Format output (json/table)
- `requirePackage()` - Validate package name

## Testing

After implementing a command:
```bash
# Build and test
make build
./gpd <command> --help
./gpd <command> [flags]
```

## Completion Checklist

- [ ] kong_config.go - Config commands
- [ ] kong_publish.go - Publish commands  
- [ ] kong_reviews.go - Reviews commands
- [ ] kong_vitals.go - Vitals commands
- [ ] kong_analytics.go - Analytics commands
- [ ] kong_purchases.go - Purchases commands
- [ ] kong_monetization.go - Monetization commands
- [ ] kong_permissions.go - Permissions commands
- [ ] kong_recovery.go - Recovery commands
- [ ] kong_apps.go - Apps commands
- [ ] kong_games.go - Games commands
- [ ] kong_integrity.go - Integrity commands
- [ ] kong_migrate.go - Migrate commands
- [ ] kong_customapp.go - CustomApp commands
- [ ] kong_grouping.go - Grouping commands

## Cleanup After Migration

Once all commands are migrated:

1. Remove Cobra from go.mod: `go mod tidy`
2. Delete old Cobra command files
3. Remove `internal/cli/cli.go` and related Cobra infrastructure
4. Update documentation
