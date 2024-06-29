package common

type WorkspaceCleanOptions struct {
	ExcludeOutputDir string
}

type MakeArtifactOptions struct {
	OutputDir      string
	OutputDirClear bool
	UseHardLink    bool
	Inspection     bool
}

const (
	OptionNameOs       = "_os"
	OptionNameArch     = "_arch"
	OptionNameExecutor = "_executor"
	OptionNameHostname = "_hostname"
	OptionNameUsername = "_username"
)
