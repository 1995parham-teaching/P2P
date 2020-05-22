package tracker

import "github.com/spf13/cobra"

func Register(root *cobra.Command) {
	root.AddCommand(
		&cobra.Command{
			Use:   "tracker",
			Short: "",
			Long:  "",
			Run: func(cmd *cobra.Command, args []string) {

			},
		},
	)
}
