package core

// region projectOptionEntity

type projectOptionEntity struct {
	Items     map[string]any
	evaluator *Evaluator
}

func newProjectOptionEntity(context *appContext, setting *projectSetting) (*projectOptionEntity, error) {
	items := context.option.GenericItems.copy()
	for i := 0; i < len(setting.Option.Items); i++ {
		declare := setting.Option.Items[i]
		result, err := context.option.findResult(setting.Name, declare)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("find option result error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Dir),
				kv("optionName", declare.Name),
			)
		}
		items[declare.Name] = result.ParsedValue
	}

	evaluator := context.evaluator.SetRootData("options", items)
	for i := 0; i < len(setting.Option.Checks); i++ {
		check := setting.Option.Checks[i]
		result, err := evaluator.EvalBoolExpr(check.expr)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("eval check error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Dir),
				kv("check", check),
			)
		}
		if !result {
			return nil, errN("load project options error",
				reason("check options error"),
				kv("projectName", setting.Name),
				kv("projectPath", setting.Dir),
				kv("check", check),
			)
		}
	}

	for i := 0; i < len(setting.Option.Items); i++ {
		optionSetting := setting.Option.Items[i]
		for j := 0; j < len(optionSetting.Assigns); j++ {
			assignSetting := optionSetting.Assigns[j]
			if err := context.option.addAssign(setting.Name, optionSetting.Name, assignSetting); err != nil {
				return nil, errW(err, "load project options error",
					reason("add option assign error"),
					kv("projectName", setting.Name),
					kv("projectPath", setting.Dir),
					kv("optionName", optionSetting.Name),
					kv("assignProject", assignSetting.Project),
					kv("assignOption", assignSetting.Option),
				)
			}
		}
	}

	option := &projectOptionEntity{
		Items:     items,
		evaluator: evaluator,
	}
	return option, nil
}

func (e *projectOptionEntity) inspect() *ProjectOptionEntityInspection {
	return newProjectOptionEntityInspection(e.Items)
}

// endregion
