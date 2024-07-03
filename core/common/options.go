package common

import . "github.com/orz-dsh/dsh/utils"

type WorkspaceCleanOptions struct {
	ExcludeOutputDir string
}

type MakeArtifactOptions struct {
	OutputDir         string
	OutputDirClear    bool
	UseHardLink       bool
	InspectSerializer Serializer
}

const (
	OptionNameCommonOs       = "_os"
	OptionNameCommonArch     = "_arch"
	OptionNameCommonExecutor = "_executor"
	OptionNameCommonHostname = "_hostname"
	OptionNameCommonUsername = "_username"
)
