package cli

import "github.com/spf13/cobra"

func addPaginationFlags(cmd *cobra.Command, all *bool) {
	cmd.Flags().BoolVar(all, "all", false, "Fetch all pages")
	cmd.Flags().BoolVar(all, "paginate", false, "Fetch all pages")
}
