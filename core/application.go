package core

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
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

func (a *Application) GetConfig() (map[string]any, error) {
	if err := a.core.LoadConfig(); err != nil {
		return nil, err
	}
	return a.core.Config.Value, nil
}

func (a *Application) MakeArtifact(options MakeArtifactOptions) (*Artifact, error) {
	artifact, err := a.core.MakeArtifact(options)
	if err != nil {
		return nil, err
	}
	return newArtifact(artifact), nil
}

func (a *Application) Inspect() (*ApplicationInspection, error) {
	return a.core.Inspect()
}

// endregion
