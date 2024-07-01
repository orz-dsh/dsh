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
	OptionNameOs       = "_os"
	OptionNameArch     = "_arch"
	OptionNameExecutor = "_executor"
	OptionNameHostname = "_hostname"
	OptionNameUsername = "_username"
)
