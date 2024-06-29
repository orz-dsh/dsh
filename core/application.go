package core

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/internal"
	. "github.com/orz-dsh/dsh/utils"
)

// region Application

type Application struct {
	core *ApplicationCore
}

func newApplication(core *ApplicationCore) *Application {
	return &Application{core: core}
}

func (a *Application) DescExtraKeyValues() KVS {
	return KVS{
		KV("core", a.core),
	}
}

func (a *Application) MakeConfigs() (map[string]any, map[string]any, error) {
	return a.core.MakeConfigs()
}

func (a *Application) MakeArtifact(options MakeArtifactOptions) (*Artifact, error) {
	artifact, err := a.core.MakeArtifact(options)
	if err != nil {
		return nil, err
	}
	return newArtifact(artifact), nil
}

// endregion
