package setup

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"oproller/config"
	"os"
	"os/exec"
)

func InitiateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init [working_space]",
		Short: "A CLI use to setup working space and clone op-geth into it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceName := args[0]
			if _, err := os.Stat(spaceName); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(spaceName, os.ModePerm)
				if err != nil {
					return err
				}
			}

			opGEth := "op-geth"
			opGEthPath := spaceName + "/" + opGEth
			if _, err := git.PlainClone(opGEthPath, false, &git.CloneOptions{
				URL:      config.OpGEthRepo,
				Progress: os.Stdout,
			}); err != nil {
				return err
			}

			cmdInit := exec.Command("go", "work", "init", opGEth)
			cmdInit.Dir = spaceName
			if err := cmdInit.Run(); err != nil {
				return err
			}

			return nil
		},
	}

}

func ClearWorkspace() *cobra.Command {
	return &cobra.Command{
		Use:   "clear [working_space]",
		Short: "A CLI use to clear working space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceName := args[0]
			if _, err := os.Stat(spaceName); errors.Is(err, os.ErrNotExist) {
				return nil
			}

			cmdRemove := exec.Command("rm", "-rf", spaceName)
			if err := cmdRemove.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}
