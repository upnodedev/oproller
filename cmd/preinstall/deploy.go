package preinstall

type ABI struct {
	DeployedBytecode struct {
		Object    string `json:"object"`
		SourceMap string `json:"sourceMap"`
	} `json:"deployedBytecode"`
}
