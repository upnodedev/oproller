package preinstall

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
	"io"
	"os"
	"os/exec"
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

	preinstallCmd.AddCommand(createPreinstallCmd())
	preinstallCmd.AddCommand(buildPreinstallCmd())
	preinstallCmd.AddCommand(registerPreinstall())
	preinstallCmd.AddCommand(runDevnet())

	return preinstallCmd
}

func createPreinstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create preinstall contract project by foundry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			preInstallName := args[0]
			preInstallName = strings.TrimSpace(preInstallName)

			if _, err := os.Stat(preInstallName); errors.Is(err, os.ErrNotExist) {
				err := os.Mkdir(preInstallName, os.ModePerm)
				if err != nil {
					return err
				}
			}

			argCmd := []string{"init", preInstallName}
			cmdBuild := exec.Command("forge", argCmd...)
			if err := cmdBuild.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}

func buildPreinstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build preinstall contract package",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdBuild := exec.Command("forge", "build")
			if err := cmdBuild.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}

func registerPreinstall() *cobra.Command {
	return &cobra.Command{
		Use:   "register [address] [preinstall_contract]",
		Short: "Register preinstall contract to optimism src",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefix := "0x"
			addrPreInstall := args[0]
			contractName := args[1]
			arrayName := strings.Split(contractName, ":")
			if len(arrayName) != 2 {
				return errors.New("preinstall_contract must be in format [contract_file.sol]:[contact_name]")
			}
			preInstallName := arrayName[1]

			if !strings.HasPrefix(addrPreInstall, prefix) {
				return errors.New("address must start with 0x")
			}
			if !common.IsHexAddress(addrPreInstall) {
				return errors.New("invalid address evm")
			}

			outPath := strings.ReplaceAll(contractName, ":", "/")
			abiFile := "out/" + outPath + ".json"

			dataAbi, err := os.ReadFile(abiFile)
			if err != nil {
				return err
			}
			var abiModel ABI
			if err := json.Unmarshal(dataAbi, &abiModel); err != nil {
				return err
			}
			hexDeployedCode := abiModel.DeployedBytecode.Object
			hexDeployedCode = strings.TrimPrefix(hexDeployedCode, "0x")

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

			// create PreinstallsExtension.sol
			t := template.Must(template.New("contract").Parse(preinstallsExtensionTemplate))
			preCompileFile := "PreinstallsExtension.sol"
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

			pathToOptimism := "../optimism"
			pathToL2Genesis := pathToOptimism + "/packages/contracts-bedrock/scripts/L2Genesis.s.sol"
			if !fileExists(pathToL2Genesis) {
				return errors.New("path_to_optimism does not exist")
			}

			pathToLibrary := pathToOptimism + "/packages/contracts-bedrock/src/libraries"
			if err := copyFile(preCompileFile, pathToLibrary+"/PreinstallsExtension.sol"); err != nil {
				return err
			}

			if err := handleRegisterL2Genesis(pathToL2Genesis); err != nil {
				return err
			}

			return nil
		},
	}
}

func runDevnet() *cobra.Command {
	return &cobra.Command{
		Use:   "devnet [up/down/clean]",
		Short: "Run devnet for testing preinstall contract package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			argCmd := args[0]
			if argCmd != "up" && argCmd != "down" && argCmd != "clean" {
				return errors.New("invalid command")
			}
			switch argCmd {
			case "down":
				cmdBuild := exec.Command("make", "devnet-down")
				if err := cmdBuild.Run(); err != nil {
					return err
				}
			case "clean":
				cmdBuild := exec.Command("make", "devnet-clean")
				if err := cmdBuild.Run(); err != nil {
					return err
				}
			default:
				cmdBuild := exec.Command("make", "devnet-up")
				if err := cmdBuild.Run(); err != nil {
					return err
				}
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
