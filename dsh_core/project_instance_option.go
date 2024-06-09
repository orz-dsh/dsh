package dsh_core

// region projectOptionInstance

type projectOptionInstance struct {
	Items     map[string]any
	evaluator *Evaluator
}

func makeProjectOption(context *appContext, setting *projectSetting) (*projectOptionInstance, error) {
	items := context.option.GenericItems.copy()
	for i := 0; i < len(setting.OptionSettings); i++ {
		declare := setting.OptionSettings[i]
		result, err := context.option.findResult(setting.Name, declare)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("find option result error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Path),
				kv("optionName", declare.Name),
			)
		}
		items[declare.Name] = result.ParsedValue
	}

	evaluator := context.evaluator.SetRootData("options", items)
	for i := 0; i < len(setting.OptionVerifySettings); i++ {
		verify := setting.OptionVerifySettings[i]
		result, err := evaluator.EvalBoolExpr(verify.expr)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("eval verify error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Path),
				kv("verify", verify),
			)
		}
		if !result {
			return nil, errN("load project options error",
				reason("verify options error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Path),
				kv("verify", verify),
			)
		}
	}

	for i := 0; i < len(setting.OptionSettings); i++ {
		optionSetting := setting.OptionSettings[i]
		for j := 0; j < len(optionSetting.AssignSettings); j++ {
			assignSetting := optionSetting.AssignSettings[j]
			if err := context.option.addAssign(setting.Name, optionSetting.Name, assignSetting); err != nil {
				return nil, errW(err, "load project options error",
					reason("add option assign error"),
					kv("projectName", setting.Name),
					kv("projectPath", setting.Path),
					kv("optionName", optionSetting.Name),
					kv("assignProject", assignSetting.Project),
					kv("assignOption", assignSetting.Option),
				)
			}
		}
	}

	option := &projectOptionInstance{
		Items:     items,
		evaluator: evaluator,
	}
	return option, nil
}

func (i *projectOptionInstance) inspect() *ProjectOptionInstanceInspection {
	return newProjectOptionInstanceInspection(i.Items)
}

// endregion

// region ProjectOptionInstanceInspection

type ProjectOptionInstanceInspection struct {
	Items map[string]any `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectOptionInstanceInspection(items map[string]any) *ProjectOptionInstanceInspection {
	return &ProjectOptionInstanceInspection{
		Items: items,
	}
}

// endregion
