package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/expr-lang/expr/vm"
)

type projectContext struct {
	workspace       *Workspace
	logger          *dsh_utils.Logger
	instanceNameMap map[string]*projectInstance
	globalOption    *globalOption
	optionLinks     map[string]*optionLink
	optionValues    map[string]string
}

func newProjectContext(workspace *Workspace, logger *dsh_utils.Logger, optionValues map[string]string) *projectContext {
	return &projectContext{
		workspace:       workspace,
		logger:          logger,
		instanceNameMap: make(map[string]*projectInstance),
		globalOption:    newGlobalOption(optionValues, logger),
		optionLinks:     make(map[string]*optionLink),
		optionValues:    make(map[string]string),
	}
}

func (context *projectContext) newProjectInstance(manifest *projectManifest, optionValues map[string]string) (*projectInstance, error) {
	if instance, exist := context.instanceNameMap[manifest.Name]; exist {
		return instance, nil
	}
	instance, err := newProjectInstance(context, manifest, optionValues)
	if err != nil {
		return nil, err
	}
	context.instanceNameMap[manifest.Name] = instance
	return instance, nil
}

type optionLink struct {
	finalTarget string
	target      string
	mapping     *vm.Program
}

func (context *projectContext) addOptionLink(sourceProject string, sourceOption string, targetProject string, targetOption string, mapping *vm.Program) error {
	sop := sourceProject + "." + sourceOption
	top := targetProject + "." + targetOption
	finalLink := &optionLink{
		finalTarget: top,
		target:      top,
		mapping:     mapping,
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

func (context *projectContext) addOptionValue(projectName string, optionName string, value string) error {
	name := projectName + "." + optionName
	if _, exist := context.optionValues[name]; exist {
		return dsh_utils.NewError("duplicate option value", map[string]interface{}{
			"name": name,
		})
	}
	context.optionValues[name] = value
	return nil
}

func (context *projectContext) getOptionLinkValue(projectName string, optionName string) (*string, bool, error) {
	name := projectName + "." + optionName
	if link, exist := context.optionLinks[name]; exist {
		if value, exist := context.optionValues[link.target]; exist {
			if link.mapping != nil {
				result, err := dsh_utils.EvalExprReturnString(link.mapping, context.globalOption.mergeItems(map[string]any{
					"value": value,
				}))
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
