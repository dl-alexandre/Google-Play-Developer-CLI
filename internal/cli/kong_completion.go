// Package cli provides shell completion generation commands.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/config"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// CompletionCmd generates shell completion scripts.
type CompletionCmd struct {
	Shell string `arg:"" help:"Shell type (bash, zsh, fish)" enum:"bash,zsh,fish"`
}

// Run generates the completion script for the specified shell.
func (cmd *CompletionCmd) Run(globals *Globals) error {
	return generateCompletion(cmd.Shell, globals)
}

// generateCompletion generates completion scripts for bash, zsh, or fish.
func generateCompletion(shell string, globals *Globals) error {
	switch shell {
	case "bash":
		return generateBashCompletion(globals)
	case "zsh":
		return generateZshCompletion(globals)
	case "fish":
		return generateFishCompletion(globals)
	default:
		return errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("unsupported shell: %s", shell)).
			WithHint("Supported shells: bash, zsh, fish")
	}
}

// getPackageCompletions returns package names for completion.
func getPackageCompletions() []string {
	packages := []string{}

	// Add from environment variable
	if pkg := config.GetEnvPackage(); pkg != "" {
		packages = append(packages, pkg)
	}

	// Add from config file
	cfg, err := config.Load()
	if err == nil && cfg.DefaultPackage != "" {
		// Only add if not already in list
		found := false
		for _, p := range packages {
			if p == cfg.DefaultPackage {
				found = true
				break
			}
		}
		if !found {
			packages = append(packages, cfg.DefaultPackage)
		}
	}

	return packages
}

// generateBashCompletion generates bash completion script.
func generateBashCompletion(_ *Globals) error {
	packages := getPackageCompletions()
	packageWords := strings.Join(packages, " ")

	script := fmt.Sprintf(`#!/bin/bash
# gpd bash completion
# Source this file: source <(gpd completion bash)
# Or add to ~/.bashrc: eval "$(gpd completion bash)"

_gpd_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # Global flags and commands
    opts="auth config publish reviews vitals analytics purchases monetization permissions recovery apps games integrity migrate customapp grouping version completion --help --package --output --pretty --timeout --store-tokens --fields --quiet --verbose --key-path --profile"
    
    # Complete based on previous word
    case "${prev}" in
        --package|-p)
            COMPREPLY=( $(compgen -W "%s" -- "${cur}") )
            return 0
            ;;
        --output|-o)
            COMPREPLY=( $(compgen -W "json table markdown csv excel" -- "${cur}") )
            return 0
            ;;
        --store-tokens)
            COMPREPLY=( $(compgen -W "auto never secure" -- "${cur}") )
            return 0
            ;;
        --timeout)
            COMPREPLY=( $(compgen -W "30s 1m 5m 10m" -- "${cur}") )
            return 0
            ;;
        --key-path)
            COMPREPLY=( $(compgen -f -- "${cur}") )
            return 0
            ;;
        auth)
            opts="login logout status check"
            ;;
        config)
            opts="init doctor path get set print export import completion"
            ;;
        publish)
            opts="upload release rollout promote halt rollback status tracks capabilities listing details images assets deobfuscation testers builds beta-groups internal-share"
            ;;
        reviews)
            opts="list get reply"
            ;;
        vitals)
            opts="crashes anrs errors ratings"  
            ;;
        analytics)
            opts="installs uninstalls ratings"
            ;;
        purchases)
            opts="verify acknowledge"
            ;;
        monetization)
            opts="subscriptions inapp"
            ;;
        integrity)
            opts="decode verify"
            ;;
        apps)
            opts="list search"
            ;;
        games)
            opts="achievements leaderboards"
            ;;
        migrate)
            opts="check plan"
            ;;
        completion)
            opts="bash zsh fish"
            ;;
    esac
    
    # Second-level command completion
    if [[ ${COMP_CWORD} -gt 1 ]]; then
        local cmd="${COMP_WORDS[1]}"
        local subcmd="${COMP_WORDS[2]}"
        
        # Complete publish upload flags
        if [[ "${cmd}" == "publish" && "${subcmd}" == "upload" ]]; then
            if [[ "${cur}" == -* ]]; then
                opts="--track --edit-id --obb-main --obb-patch --obb-main-ref-version --obb-patch-ref-version --no-auto-commit --dry-run"
            fi
        fi
        
        # Complete track options for publish commands
        if [[ "${prev}" == "--track" ]]; then
            COMPREPLY=( $(compgen -W "internal alpha beta production" -- "${cur}") )
            return 0
        fi
    fi
    
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
    return 0
}

complete -F _gpd_completions gpd
`, packageWords)

	_, _ = fmt.Fprint(os.Stdout, script)
	return nil
}

// generateZshCompletion generates zsh completion script.
func generateZshCompletion(_ *Globals) error {
	packages := getPackageCompletions()

	packageCompletions := ""
	for _, p := range packages {
		packageCompletions += fmt.Sprintf("        '%s'\n", p)
	}

	script := fmt.Sprintf(`#compdef gpd
# gpd zsh completion
# To use: source <(gpd completion zsh)
# Or add to ~/.zshrc: eval "$(gpd completion zsh)"

_gpd() {
    local -a commands
    commands=(
        'auth:Authentication commands'
        'config:Configuration commands'
        'publish:Publishing commands'
        'reviews:Review management commands'
        'vitals:Android vitals commands'
        'analytics:Analytics commands'
        'purchases:Purchase verification commands'
        'monetization:Monetization commands'
        'permissions:Permissions management'
        'recovery:App recovery commands'
        'apps:App discovery commands'
        'games:Google Play Games services'
        'integrity:Play Integrity API commands'
        'migrate:Migration commands'
        'customapp:Custom app publishing'
        'grouping:App access grouping'
        'version:Show version information'
        'completion:Generate shell completion scripts'
    )
    
    local -a output_formats
    output_formats=('json' 'table' 'markdown' 'csv' 'excel')
    
    local -a track_names
    track_names=('internal' 'alpha' 'beta' 'production')
    
    local -a store_tokens
    store_tokens=('auto' 'never' 'secure')
    
    local -a package_names
    package_names=(
%s    )
    
    _arguments -C \
        '(-p --package)'{-p,--package}'[App package name]:package:_gpd_packages' \
        '(-o --output)'{-o,--output}'[Output format]:format:_gpd_output_formats' \
        '--pretty[Pretty print JSON output]' \
        '--timeout[Network timeout]:timeout:' \
        '--store-tokens[Token storage]:storage:_gpd_store_tokens' \
        '--fields[JSON field projection]:fields:' \
        '--quiet[Suppress non-error output]' \
        '(-v --verbose)'{-v,--verbose}'[Enable verbose logging]' \
        '--key-path[Path to service account key file]:path:_files' \
        '--profile[Configuration profile to use]:profile:' \
        '1: :->command' \
        '*::arg:->args'
    
    case "$state" in
        command)
            _describe -t commands 'gpd command' commands
            ;;
        args)
            case "$line[1]" in
                config)
                    local -a config_cmds
                    config_cmds=(init doctor path get set print export import completion)
                    _describe -t commands 'config command' config_cmds
                    ;;
                auth)
                    local -a auth_cmds
                    auth_cmds=(login logout status check)
                    _describe -t commands 'auth command' auth_cmds
                    ;;
                publish)
                    local -a publish_cmds
                    publish_cmds=(upload release rollout promote halt rollback status tracks capabilities listing details images assets deobfuscation testers builds beta-groups internal-share)
                    _describe -t commands 'publish command' publish_cmds
                    ;;
                reviews)
                    local -a reviews_cmds
                    reviews_cmds=(list get reply)
                    _describe -t commands 'reviews command' reviews_cmds
                    ;;
                completion)
                    local -a shells
                    shells=(bash zsh fish)
                    _describe -t commands 'shell' shells
                    ;;
                *)
                    _files
                    ;;
            esac
            ;;
    esac
}

_gpd_output_formats() {
    local -a formats
    formats=('json' 'table' 'markdown' 'csv' 'excel')
    _describe -t formats 'output format' formats
}

_gpd_store_tokens() {
    local -a modes
    modes=('auto' 'never' 'secure')
    _describe -t modes 'token storage' modes
}

_gpd_packages() {
    local -a packages
    packages=(
%s    )
    _describe -t packages 'package' packages
}

compdef _gpd gpd
`, packageCompletions, packageCompletions)

	_, _ = fmt.Fprint(os.Stdout, script)
	return nil
}

// generateFishCompletion generates fish completion script.
func generateFishCompletion(_ *Globals) error {
	packages := getPackageCompletions()

	var sb strings.Builder

	// Header
	sb.WriteString("# gpd fish completion\n")
	sb.WriteString("# To use: gpd completion fish | source\n")
	sb.WriteString("# Or save to: gpd completion fish > ~/.config/fish/completions/gpd.fish\n\n")

	// Disable file completions by default
	sb.WriteString("complete -c gpd -f\n\n")

	// Global flags
	sb.WriteString("# Global flags\n")
	sb.WriteString("complete -c gpd -l package -s p -d 'App package name'")
	for _, p := range packages {
		_, _ = fmt.Fprintf(&sb, " -a '%s'", p)
	}
	sb.WriteString("\n")

	sb.WriteString("complete -c gpd -l output -s o -d 'Output format' -a 'json table markdown csv excel'\n")
	sb.WriteString("complete -c gpd -l pretty -d 'Pretty print JSON output'\n")
	sb.WriteString("complete -c gpd -l timeout -d 'Network timeout'\n")
	sb.WriteString("complete -c gpd -l store-tokens -d 'Token storage' -a 'auto never secure'\n")
	sb.WriteString("complete -c gpd -l fields -d 'JSON field projection'\n")
	sb.WriteString("complete -c gpd -l quiet -d 'Suppress non-error output'\n")
	sb.WriteString("complete -c gpd -l verbose -s v -d 'Enable verbose logging'\n")
	sb.WriteString("complete -c gpd -l key-path -d 'Path to service account key file' -F\n")
	sb.WriteString("complete -c gpd -l profile -d 'Configuration profile to use'\n\n")

	// Commands
	sb.WriteString("# Commands\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a auth -d 'Authentication commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a config -d 'Configuration commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a publish -d 'Publishing commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a reviews -d 'Review management commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a vitals -d 'Android vitals commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a analytics -d 'Analytics commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a purchases -d 'Purchase verification commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a monetization -d 'Monetization commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a permissions -d 'Permissions management'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a recovery -d 'App recovery commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a apps -d 'App discovery commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a games -d 'Google Play Games services'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a integrity -d 'Play Integrity API commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a migrate -d 'Migration commands'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a customapp -d 'Custom app publishing'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a grouping -d 'App access grouping'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a version -d 'Show version information'\n")
	sb.WriteString("complete -c gpd -n '__fish_use_subcommand' -a completion -d 'Generate shell completion scripts'\n\n")

	// Config subcommands
	sb.WriteString("# Config subcommands\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a init -d 'Initialize project configuration'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a doctor -d 'Diagnose configuration issues'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a path -d 'Show configuration paths'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a get -d 'Get configuration value'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a set -d 'Set configuration value'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a print -d 'Print configuration'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a export -d 'Export configuration to file'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a import -d 'Import configuration from file'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from config' -a completion -d 'Generate shell completion script'\n\n")

	// Auth subcommands
	sb.WriteString("# Auth subcommands\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from auth' -a login -d 'Authenticate with Google Play'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from auth' -a logout -d 'Sign out and remove credentials'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from auth' -a status -d 'Check authentication status'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from auth' -a check -d 'Validate credentials'\n\n")

	// Publish subcommands
	sb.WriteString("# Publish subcommands\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a upload -d 'Upload APK or AAB'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a release -d 'Create or update a release'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a rollout -d 'Update rollout percentage'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a promote -d 'Promote between tracks'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a halt -d 'Halt a rollout'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a rollback -d 'Rollback to previous version'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a status -d 'Get track status'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish' -a tracks -d 'List all tracks'\n\n")

	// Track completions for publish commands
	sb.WriteString("# Track completions\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish; and __fish_seen_subcommand_from upload' -l track -d 'Target track' -a 'internal alpha beta production'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish; and __fish_seen_subcommand_from release' -l track -d 'Release track' -a 'internal alpha beta production'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from publish; and __fish_seen_subcommand_from rollout' -l track -d 'Release track' -a 'internal alpha beta production'\n\n")

	// Completion subcommand
	sb.WriteString("# Completion subcommand shells\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from completion' -a bash -d 'Generate bash completion'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from completion' -a zsh -d 'Generate zsh completion'\n")
	sb.WriteString("complete -c gpd -n '__fish_seen_subcommand_from completion' -a fish -d 'Generate fish completion'\n")

	_, _ = fmt.Fprint(os.Stdout, sb.String())
	return nil
}
