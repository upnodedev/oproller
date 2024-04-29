# OP-Roller (oproller)

This is a simple tool to setup work spaces for development custom precompile and predeploy from op-geth

## Requirements

- Install go version 1.20 or higher

## How to install

- Clone this repository
- Run `go install`

## How to use

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
oproller precompile new my-precompile 0x123
```

- Step3: Build the precompile. Ensure that you  go to the workspace directory
```shell
oproller precompile build
```


After that you are able to see the precompile file in the workspace directory. And then start developing your precompile.

### Note:
Please do not edit the package name and the module of precompile file. It will be used to register into to the op-geth.

