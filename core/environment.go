package core

import (
	. "github.com/orz-dsh/dsh/core/internal"
	. "github.com/orz-dsh/dsh/utils"
)

type Environment struct {
	core *EnvironmentCore
}

func NewEnvironment(logger *Logger, variables map[string]string) (*Environment, error) {
	core, err := NewEnvironmentCore(logger, variables)
	if err != nil {
		return nil, err
	}
	environment := &Environment{
		core: core,
	}
	return environment, nil
}

func (e *Environment) DescExtraKeyValues() KVS {
	return KVS{
		KV("core", e.core),
	}
}
