package cli

import (
	"context"
	"fmt"

	"github.com/dl-alexandre/gpd/internal/output"
)

// AuthCmd contains authentication commands.
type AuthCmd struct {
	Status AuthStatusCmd `cmd:"" help:"Check authentication status"`
	Login  AuthLoginCmd  `cmd:"" help:"Authenticate with Google Play"`
	Logout AuthLogoutCmd `cmd:"" help:"Sign out and clear credentials"`
}

// AuthStatusCmd checks authentication status.
type AuthStatusCmd struct{}

// Run executes the auth status command.
func (cmd *AuthStatusCmd) Run(globals *Globals) error {
	ctx := context.Background()
	authMgr := newAuthManager()
	status, err := authMgr.GetStatus(ctx)
	if err != nil {
		return err
	}

	result := output.NewResult(status)
	return outputResult(result, globals.Output, globals.Pretty)
}

// AuthLoginCmd authenticates with Google Play.
type AuthLoginCmd struct {
	Key string `help:"Path to service account key file" type:"existingfile"`
}

// Run executes the auth login command.
func (cmd *AuthLoginCmd) Run(globals *Globals) error {
	ctx := context.Background()
	authMgr := newAuthManager()

	_, err := authMgr.Authenticate(ctx, cmd.Key)
	if err != nil {
		return err
	}

	fmt.Println("Authentication successful")
	return nil
}

// AuthLogoutCmd signs out and clears credentials.
type AuthLogoutCmd struct{}

// Run executes the auth logout command.
func (cmd *AuthLogoutCmd) Run(globals *Globals) error {
	authMgr := newAuthManager()
	authMgr.Clear()

	fmt.Println("Signed out successfully")
	return nil
}
