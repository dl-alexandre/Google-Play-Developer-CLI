package cli

import (
	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// Permissions Commands
// ============================================================================

// PermissionsCmd contains permissions management commands.
type PermissionsCmd struct {
	Users  PermissionsUsersCmd  `cmd:"" help:"Manage users"`
	Grants PermissionsGrantsCmd `cmd:"" help:"Manage grants"`
	List   PermissionsListCmd   `cmd:"" help:"List permissions"`
}

// PermissionsUsersCmd manages users.
type PermissionsUsersCmd struct {
	Add    PermissionsUsersAddCmd    `cmd:"" help:"Add a user"`
	Remove PermissionsUsersRemoveCmd `cmd:"" help:"Remove a user"`
	List   PermissionsUsersListCmd   `cmd:"" help:"List users"`
}

// PermissionsUsersAddCmd adds a user.
type PermissionsUsersAddCmd struct {
	Email string `help:"User email address" required:""`
	Role  string `help:"User role" required:"" enum:"admin,developer,viewer"`
}

// Run executes the users add command.
func (cmd *PermissionsUsersAddCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "permissions users add not yet implemented")
}

// PermissionsUsersRemoveCmd removes a user.
type PermissionsUsersRemoveCmd struct {
	Email string `help:"User email address" required:""`
}

// Run executes the users remove command.
func (cmd *PermissionsUsersRemoveCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "permissions users remove not yet implemented")
}

// PermissionsUsersListCmd lists users.
type PermissionsUsersListCmd struct{}

// Run executes the users list command.
func (cmd *PermissionsUsersListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "permissions users list not yet implemented")
}

// PermissionsGrantsCmd manages grants.
type PermissionsGrantsCmd struct {
	Add    PermissionsGrantsAddCmd    `cmd:"" help:"Add a grant"`
	Remove PermissionsGrantsRemoveCmd `cmd:"" help:"Remove a grant"`
	List   PermissionsGrantsListCmd   `cmd:"" help:"List grants"`
}

// PermissionsGrantsAddCmd adds a grant.
type PermissionsGrantsAddCmd struct {
	Email  string `help:"User email address" required:""`
	Grant  string `help:"Permission grant" required:""`
	Expiry string `help:"Grant expiry date (YYYY-MM-DD)"`
}

// Run executes the grants add command.
func (cmd *PermissionsGrantsAddCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "permissions grants add not yet implemented")
}

// PermissionsGrantsRemoveCmd removes a grant.
type PermissionsGrantsRemoveCmd struct {
	Email string `help:"User email address" required:""`
	Grant string `help:"Permission grant" required:""`
}

// Run executes the grants remove command.
func (cmd *PermissionsGrantsRemoveCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "permissions grants remove not yet implemented")
}

// PermissionsGrantsListCmd lists grants.
type PermissionsGrantsListCmd struct {
	Email string `help:"Filter by user email"`
}

// Run executes the grants list command.
func (cmd *PermissionsGrantsListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "permissions grants list not yet implemented")
}

// PermissionsListCmd lists permissions.
type PermissionsListCmd struct{}

// Run executes the permissions list command.
func (cmd *PermissionsListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "permissions list not yet implemented")
}

// ============================================================================
// Recovery Commands
// ============================================================================

// RecoveryCmd contains app recovery commands.
type RecoveryCmd struct {
	List   RecoveryListCmd   `cmd:"" help:"List recovery actions"`
	Create RecoveryCreateCmd `cmd:"" help:"Create recovery action"`
	Deploy RecoveryDeployCmd `cmd:"" help:"Deploy recovery"`
	Cancel RecoveryCancelCmd `cmd:"" help:"Cancel recovery"`
}

// RecoveryListCmd lists recovery actions.
type RecoveryListCmd struct {
	Status string `help:"Filter by status: pending,active,completed,cancelled,failed"`
}

// Run executes the recovery list command.
func (cmd *RecoveryListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "recovery list not yet implemented")
}

// RecoveryCreateCmd creates a recovery action.
type RecoveryCreateCmd struct {
	Type   string `help:"Recovery type" required:"" enum:"rollback,emergency_update,version_hold"`
	Target string `help:"Target version or track"`
	Reason string `help:"Reason for recovery" required:""`
}

// Run executes the recovery create command.
func (cmd *RecoveryCreateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "recovery create not yet implemented")
}

// RecoveryDeployCmd deploys a recovery.
type RecoveryDeployCmd struct {
	ID string `arg:"" help:"Recovery action ID" required:""`
}

// Run executes the recovery deploy command.
func (cmd *RecoveryDeployCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "recovery deploy not yet implemented")
}

// RecoveryCancelCmd cancels a recovery.
type RecoveryCancelCmd struct {
	ID     string `arg:"" help:"Recovery action ID" required:""`
	Reason string `help:"Reason for cancellation"`
}

// Run executes the recovery cancel command.
func (cmd *RecoveryCancelCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "recovery cancel not yet implemented")
}

// ============================================================================
// Integrity Commands
// ============================================================================

// IntegrityCmd contains Play Integrity API commands.
type IntegrityCmd struct {
	Decode IntegrityDecodeCmd `cmd:"" help:"Decode integrity token"`
}

// IntegrityDecodeCmd decodes an integrity token.
type IntegrityDecodeCmd struct {
	Token  string `arg:"" help:"Integrity token to decode" required:""`
	Verify bool   `help:"Verify token signature"`
}

// Run executes the integrity decode command.
func (cmd *IntegrityDecodeCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "integrity decode not yet implemented")
}
