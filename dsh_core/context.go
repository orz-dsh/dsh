package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/expr-lang/expr/vm"
)

type Context struct {
	Workspace       *Workspace
	Logger          *dsh_utils.Logger
	OptionSelector  *OptionSelector
	Project         *Project
	instanceNameMap map[string]*projectInstance
	optionLinks     map[string]*optionLink
	optionValues    map[string]string
}

type OptionSelector struct {
}

func NewContext(workspace *Workspace, logger *dsh_utils.Logger) *Context {
	return &Context{
		Workspace:       workspace,
		Logger:          logger,
		OptionSelector:  &OptionSelector{},
		instanceNameMap: make(map[string]*projectInstance),
		optionLinks:     make(map[string]*optionLink),
		optionValues:    make(map[string]string),
	}
}

func (context *Context) newProjectInstance(info *projectInfo, optionValues map[string]string) (*projectInstance, error) {
	if instance, exist := context.instanceNameMap[info.name]; exist {
		return instance, nil
	}
	instance, err := newProjectInstance(context, info, optionValues)
	if err != nil {
		return nil, err
	}
	context.instanceNameMap[info.name] = instance
	return instance, nil
}

type optionLink struct {
	finalTarget string
	target      string
	mapper      *vm.Program
}

func (context *Context) addOptionLink(sourceProject string, sourceOption string, targetProject string, targetOption string, mapper *vm.Program) error {
	sop := sourceProject + "." + sourceOption
	top := targetProject + "." + targetOption
	finalLink := &optionLink{
		target: top,
		mapper: mapper,
	}
	if topLink, exist := context.optionLinks[top]; exist {
		finalLink.finalTarget = topLink.finalTarget
	}
	if link, exist := context.optionLinks[sop]; exist {
		if link.finalTarget != finalLink.finalTarget {
			return dsh_utils.NewError("option link conflict", map[string]interface{}{
				"source":  sop,
				"target1": finalLink,
				"target2": link,
			})
		}
	} else {
		context.optionLinks[sop] = finalLink
	}
	return nil
}

func (context *Context) addOptionValue(projectName string, optionName string, value string) error {
	name := projectName + "." + optionName
	if _, exist := context.optionValues[name]; exist {
		return dsh_utils.NewError("duplicate option value", map[string]interface{}{
			"name": name,
		})
	}
	context.optionValues[name] = value
	return nil
}

func (context *Context) getOptionLinkValue(projectName string, optionName string) (*string, bool, error) {
	name := projectName + "." + optionName
	if link, exist := context.optionLinks[name]; exist {
		if value, exist := context.optionValues[link.target]; exist {
			if link.mapper != nil {
				result, err := dsh_utils.EvalExprReturnString(link.mapper, map[string]any{
					"value": value,
				})
				if err != nil {
					return nil, false, err
				}
				return result, true, nil
			}
			return &value, true, nil
		}
	}
	return nil, false, nil
}
