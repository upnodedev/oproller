package precompile

import (
	"errors"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"text/template"
	"unicode"
)

var precompileTemplate = `
package {{.PackageName}}

type {{.StructName}} struct{}

func (c *{{.StructName}}) RequiredGas(input []byte) uint64 {
	// TODO: implement RequiredGas
	panic("implement me")
}

func (c *{{.StructName}}) Run(input []byte) ([]byte, error) {
	// TODO: implement Run
	panic("implement me")
}
`

func Cmd() *cobra.Command {
	precompileCmd := &cobra.Command{
		Use:   "precompile",
		Short: "All commands related to precompile contracts",
	}

	precompileCmd.AddCommand(newCmd())
	precompileCmd.AddCommand(buildCmd())

	return precompileCmd
}

func buildCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build precompile contract package",
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat("bin"); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir("bin", os.ModePerm)
				if err != nil {
					return err
				}
			}

			cmdBuild := exec.Command("make", "geth")
			cmdBuild.Dir = "./op-geth"
			if err := cmdBuild.Run(); err != nil {
				return err
			}

			cmdCopy := exec.Command("cp", "./op-geth/build/bin/geth", "./bin/")
			if err := cmdCopy.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}

func newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new [name] [address]",
		Short: "Register a precompile contract package",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			preCompileName := args[0]

			runes := []rune(preCompileName)
			if unicode.IsLetter(runes[0]) {
				runes[0] = unicode.ToUpper(runes[0])
			}
			structName := string(runes)
			data := struct {
				PackageName string
				StructName  string
			}{
				PackageName: preCompileName,
				StructName:  structName,
			}

			// mkdir precompile package
			if _, err := os.Stat(preCompileName); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(preCompileName, os.ModePerm)
				if err != nil {
					return err
				}
			}

			// init go mod
			if _, err := os.Stat(preCompileName + "/go.mod"); errors.Is(err, os.ErrNotExist) {
				cmdInit := exec.Command("go", "mod", "init", preCompileName)
				cmdInit.Dir = preCompileName
				if err := cmdInit.Run(); err != nil {
					return err
				}
			}

			// add go work
			cmdAddGoWork := exec.Command("go", "work", "use", preCompileName)
			if err := cmdAddGoWork.Run(); err != nil {
				return err
			}

			// create precompile.go
			t := template.Must(template.New("package").Parse(precompileTemplate))
			preCompileFile := preCompileName + "/" + preCompileName + ".go"
			outputFile, err := os.Create(preCompileFile)
			if err != nil {
				return err
			}
			defer outputFile.Close()

			// Execute the template and write the output to the file
			err = t.Execute(outputFile, data)
			if err != nil {
				return err
			}

			h := NewHandle(args[1], data.StructName, data.PackageName)
			if err := h.registerPrecompile("./op-geth"); err != nil {
				return err
			}

			return nil
		},
	}
}
