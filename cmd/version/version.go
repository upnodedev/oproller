package version

import (
	"fmt"
	"github.com/spf13/cobra"
	"oproller/version"
)

func Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of op-roller",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("💈 OP Roller version", version.BuildVersion)
			fmt.Println("💈 Build time:", version.BuildTime)
			fmt.Println("💈 Git commit:", version.BuildCommit)
		},
	}
}
