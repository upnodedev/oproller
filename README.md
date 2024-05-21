# OP-Roller (oproller)

This is a simple tool to setup work spaces for development custom precompile and predeploy from op-geth

## Requirements

- Install go version 1.20 or higher

## How to install

- Clone this repository
- Run `make install`

## How to add new precompile

### Add new precompile

- Step1: Create new workspace
```shell
oproller init <workspace-name>
```
```shell
oproller init my-workspace
```

- Step2: Add new precompile we need go into the workspace directory
```shell
cd my-workspace
oproller precompile new <precompile-name> <address-of-precompile>
```
```shell
oproller precompile new my-precompile 0x1234
```

- Step3: Build the precompile. Ensure that you  go to the workspace directory
```shell
oproller precompile build
```


After that you are able to see the precompile file in the workspace directory. And then start developing your precompile.

### Clear up the workspace
```shell
 oproller clear <workspace-name>
```
```shell
 oproller clear my-workspace
```



### Note:
Please do not edit the package name and the module of precompile file. It will be used to register into to the op-geth.

## How to add new preinstall

### Add new preinstall

To develop the preinstall, we need to create the smart contract project to develop the preinstall. The preinstall is the smart contract that will be deployed to the blockchain to support the precompile.
We can use the foundry or hardhat to create the project. After that we need to build the project and get the deploy bytecode of the smart contract. When we have the deploy bytecode, we can add the preinstall to the optimism following the steps below:

- Step1: Generate the preinstall extension
```shell
oproller preinstall generate [name] [address] [hex_deployed_code]
```
- name: The name of the preinstall
- address: The address of the preinstall
- hex_deployed_code: The hex of the deployed code of the preinstall

- Step2: Register the preinstall to the optimism project
```shell
oproller preinstall register [path_to_preinstall_contract] [path_to_optimism]
```

- path_to_preinstall_contract: The path to the preinstall contract that you have generated from the step1
- path_to_optimism: The path to the optimism project that you want to add the preinstall