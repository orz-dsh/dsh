package core

import (
	. "github.com/orz-dsh/dsh/core/internal"
	. "github.com/orz-dsh/dsh/utils"
)

// region Artifact

type Artifact struct {
	core *ArtifactCore
}

func newArtifact(core *ArtifactCore) *Artifact {
	return &Artifact{core: core}
}

func (a *Artifact) DescExtraKeyValues() KVS {
	return KVS{
		KV("core", a.core),
	}
}

func (a *Artifact) GetOutputDir() string {
	return a.core.OutputDir
}

func (a *Artifact) ExecuteInChildProcess(targetGlob string) (int, error) {
	return a.core.ExecuteInChildProcess(targetGlob)
}

func (a *Artifact) ExecuteInThisProcess(targetGlob string) error {
	return a.core.ExecuteInThisProcess(targetGlob)
}

// endregion
