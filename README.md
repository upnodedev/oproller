# OP-Roller (oproller)

This is a simple tool to setup work spaces for development custom precompile and predeploy from op-geth/optimism.

## Requirements

- Install go version 1.20 or higher
[Install go](https://golang.org/doc/install)
- Docker and docker-compose 
[Install docker](https://docs.docker.com/get-docker/)
- Install foundry
[Install foundry](https://book.getfoundry.sh/getting-started/installation)

## How to install oproller

- Clone this repository
```shell
git clone https://github.com/johnyupnode/oproller.git
```
- Run `make install`
```shell
cd oproller
make install
```

## How to add new precompile

### Add new precompile

- Step1: Create new workspace
```shell
oproller init precompile <workspace-name>
```
```shell
oproller init precompile my-workspace
```
After that you will see the workspace directory in the current directory. The structure of the workspace directory is as below:
```
my-workspace:
    - precompile
        - op-geth
            - ...
        - go.work
```

- Step2: Add new precompile we need go into the workspace directory
```shell
cd my-workspace
oproller precompile new <precompile-name> <address-of-precompile>
```
```shell
oproller precompile new my-precompile 0xE11049cf6DFeB008e198d9c1155aEaA35b2e2Ba2
```

After that you will see the precompile file in the workspace directory. The structure of the workspace directory is as below:
```
my-workspace:
    - precompile
        - my-precompile
            - ...
            - go.mod
        - op-geth
            - ...
            - go.mod
        - go.work
```

- Step3: Build the precompile. Ensure that you  go to the precompile workspace directory (cd my-workspace/precompile)
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

- Step1: Create new workspace
```shell
oproller init preinstall <workspace-name>
```
```shell
oproller init preinstall my-workspace
```
After that you will see the workspace directory in the current directory. The structure of the workspace directory is as below:
```
my-workspace:
    - preinstall
        - optimism
            - ...
```

- Step2: Add new preinstall we need go into the workspace directory
```shell
cd my-workspace
oproller preinstall create <preinstall-name>
```
```shell
oproller preinstall create my-preinstall
```

This command will use foundry to create a new preinstall smartcontract template. After that you will see the preinstall file in the workspace directory. The structure of the workspace directory is as below:
```
my-workspace:
    - preinstall
        - optimism
            - ...
        - my-preinstall
            - ...
```

- Step3: Build the preinstall. Ensure that you  go to the project of preinstall workspace directory (cd my-workspace/preinstall/my-preinstall)
```shell
cd my-workspace/preinstall/my-preinstall
oproller preinstall build
```

- Step4: Register the preinstall into the preinstall list of the optimism. Ensure that you  go to the project of preinstall workspace directory (cd my-workspace/preinstall/my-preinstall)
```shell
cd my-workspace/preinstall/my-preinstall
oproller preinstall register <address-of-preinstall> <contract-file:contract-name>
```
```shell
oproller preinstall register 0xE11049cf6DFeB008e198d9c1155aEaA35b2e2Ba2 Counter.sol:Counter
```

- Step5: Deploy devnet to test the preinstall. Ensure that you  go to the project of optimism workspace directory (cd my-workspace/preinstall/optimism)
```shell
cd my-workspace/preinstall/optimism
oproller preinstall devnet <command>
```
The <command> can be `up`, `down`, `clean`. The `up` command will deploy the devnet, `down` command will stop the devnet, `clean` command will remove the devnet. 

```shell
oproller preinstall devnet up
```
