package dsh_core

import "dsh/dsh_utils"

type Context struct {
	Workspace       *Workspace
	Logger          *dsh_utils.Logger
	OptionSelector  *OptionSelector
	Project         *Project
	instanceNameMap map[string]*projectInstance
}

type OptionSelector struct {
}

func NewContext(workspace *Workspace, logger *dsh_utils.Logger) *Context {
	return &Context{
		Workspace:       workspace,
		Logger:          logger,
		OptionSelector:  &OptionSelector{},
		instanceNameMap: make(map[string]*projectInstance),
	}
}

func (context *Context) newProjectInstance(info *projectInfo) (*projectInstance, error) {
	if instance, exist := context.instanceNameMap[info.name]; exist {
		return instance, nil
	}
	instance, err := newProjectInstance(context, info)
	if err != nil {
		return nil, err
	}
	context.instanceNameMap[info.name] = instance
	return instance, nil
}
