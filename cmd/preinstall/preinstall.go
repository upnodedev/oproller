package preinstall

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"regexp"
	"strings"
	"text/template"
)

var preinstallsExtensionTemplate = `
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

library PreinstallsExtension {
	address internal constant {{.Name}} = {{.Address}};
	bytes internal constant {{.Name}}Code = hex"{{.HexDeployedCode}}";

function getDeployedCode(address _addr, uint256 _chainID) internal pure returns (bytes memory out_) {
        if (_addr == {{.Name}}) return {{.Name}}Code;

        revert("PreinstallsExtension: unknown preinstall");
    }

    /// @notice Returns the name of the preinstall at the given address.
    function getName(address _addr) internal pure returns (string memory out_) {
        if (_addr == {{.Name}}) return "{{.Name}}";

        revert("PreinstallsExtension: unnamed preinstall");
    }

    function getPreinstallAddresses() internal pure returns (address[] memory out_) {
        out_ = new address[](1);
        out_[0] = {{.Name}};
    }
}
`

func Cmd() *cobra.Command {
	preinstallCmd := &cobra.Command{
		Use:   "preinstall",
		Short: "All commands related to preinstall contracts",
	}

	preinstallCmd.AddCommand(generatePreinstall())
	preinstallCmd.AddCommand(registerPreinstall())

	return preinstallCmd
}

func generatePreinstall() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [name] [address] [hex_deployed_code]",
		Short: "Generate preinstall contract",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := "0x"
			preInstallName := args[0]
			addrPreInstall := args[1]
			hexDeployedCode := args[2]
			if !strings.HasPrefix(addrPreInstall, prefix) {
				return errors.New("address must start with 0x")
			}
			if !isValidHex(hexDeployedCode) {
				return errors.New("hex_deployed_code must be a valid hexadecimal string")
			}
			preInstallName = strings.TrimSpace(preInstallName)

			data := struct {
				Name            string
				Address         string
				HexDeployedCode string
			}{
				Name:            preInstallName,
				Address:         addrPreInstall,
				HexDeployedCode: hexDeployedCode,
			}
			// mkdir preinstall dir
			if _, err := os.Stat(data.Name); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(data.Name, os.ModePerm)
				if err != nil {
					return err
				}
			}

			// create precompile.go
			t := template.Must(template.New("contract").Parse(preinstallsExtensionTemplate))
			preCompileFile := data.Name + "/" + "PreinstallsExtension.sol"
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

			fmt.Println("Preinstall contract generate successfully! Please check the file at " + preCompileFile + " for more details.")

			return nil
		},
	}
}

func registerPreinstall() *cobra.Command {
	return &cobra.Command{
		Use:   "register [path_to_preinstall_contract] [path_to_optimism]",
		Short: "Register preinstall contract to optimism src",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pathToContract := args[0]
			pathToOptimism := args[1]
			if !fileExists(pathToContract) {
				return errors.New("path_to_preinstall_contract does not exist")
			}

			pathToL2Genesis := pathToOptimism + "/packages/contracts-bedrock/scripts/L2Genesis.s.sol"
			if !fileExists(pathToL2Genesis) {
				return errors.New("path_to_optimism does not exist")
			}

			pathToLibrary := pathToOptimism + "/packages/contracts-bedrock/src/libraries"
			if err := copyFile(pathToContract, pathToLibrary+"/PreinstallsExtension.sol"); err != nil {
				return err
			}

			if err := handleRegisterL2Genesis(pathToL2Genesis); err != nil {
				return err
			}

			return nil
		},
	}
}

func handleRegisterL2Genesis(pathToL2Genesis string) error {
	content, err := os.ReadFile(pathToL2Genesis)
	if err != nil {
		return err
	}
	strContent := string(content)

	// import PreinstallsExtension.sol
	regexImport := `(?m)^pragma solidity \d+\.\d+\.\d+;`

	replaceImport := `pragma solidity 0.8.15;

import { PreinstallsExtension } from "src/libraries/PreinstallsExtension.sol";`

	re, err := regexp.Compile(regexImport)
	if err != nil {
		return err
	}

	if !strings.Contains(strContent, "PreinstallsExtension.sol") {
		strContent = re.ReplaceAllString(strContent, replaceImport)
	}

	// add new function to L2Genesis.s.sol
	newFunc := `	/// @notice Sets the bytecode in state
    function _setPreinstallExtensionCode(address _addr) internal {
        string memory cname = PreinstallsExtension.getName(_addr);
        console.log("Setting %s preinstall extension code at: %s", cname, _addr);
        vm.etch(_addr, PreinstallsExtension.getDeployedCode(_addr, cfg.l2ChainID()));
        // during testing in a shared L1/L2 account namespace some preinstalls may already have been inserted and used.
        if (vm.getNonce(_addr) == 0) {
            vm.setNonce(_addr, 1);
        }
    }`
	if !strings.Contains(strContent, "_setPreinstallExtensionCode") {
		lastClosing := strings.LastIndex(strContent, "}")
		strContent = strContent[:lastClosing] + "\n" + newFunc + "\n}"
	}

	// setPreinstalls with new function
	if !strings.Contains(strContent, "PreinstallsExtension.getPreinstallAddresses") {
		addPreinstalls := `_setPreinstallCode(Preinstalls.BeaconBlockRoots);
	 	for (uint256 i; i < PreinstallsExtension.getPreinstallAddresses().length; i++) {
            _setPreinstallExtensionCode(PreinstallsExtension.getPreinstallAddresses()[i]);
		}`
		strContent = strings.ReplaceAll(strContent, "_setPreinstallCode(Preinstalls.BeaconBlockRoots);", addPreinstalls)
	}

	// write to L2Genesis.s.sol
	if err := os.WriteFile(pathToL2Genesis, []byte(strContent), 0644); err != nil {
		return err
	}

	return nil
}

func isValidHex(s string) bool {
	// Regular expression to match a hexadecimal string
	var re = regexp.MustCompile(`^[0-9a-fA-F]+$`)
	return re.MatchString(s)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file.", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
