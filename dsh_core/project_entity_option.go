package dsh_core

// region option

type projectOption struct {
	Items     map[string]any
	evaluator *Evaluator
}

func makeProjectOption(context *appContext, entity *projectSchema) (*projectOption, error) {
	items := context.option.GenericItems.copy()
	for i := 0; i < len(entity.OptionDeclares); i++ {
		declare := entity.OptionDeclares[i]
		result, err := context.option.findResult(entity.Name, declare)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("find option result error"),
				kv("projectName", entity.Name),
				kv("projectPath", entity.Path),
				kv("optionName", declare.Name),
			)
		}
		items[declare.Name] = result.ParsedValue
	}

	evaluator := context.evaluator.SetRootData("options", items)
	for i := 0; i < len(entity.OptionVerifies); i++ {
		verify := entity.OptionVerifies[i]
		result, err := evaluator.EvalBoolExpr(verify.expr)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("eval verify error"),
				kv("projectName", entity.Name),
				kv("projectPath", entity.Path),
				kv("verify", verify),
			)
		}
		if !result {
			return nil, errN("load project options error",
				reason("verify options error"),
				kv("projectName", entity.Name),
				kv("projectPath", entity.Path),
				kv("verify", verify),
			)
		}
	}

	for i := 0; i < len(entity.OptionDeclares); i++ {
		declare := entity.OptionDeclares[i]
		for j := 0; j < len(declare.Assigns); j++ {
			assign := declare.Assigns[j]
			if err := context.option.addAssign(entity.Name, declare.Name, assign.Project, assign.Option, assign.mapping); err != nil {
				return nil, errW(err, "load project options error",
					reason("add option assign error"),
					kv("projectName", entity.Name),
					kv("projectPath", entity.Path),
					kv("optionName", declare.Name),
					kv("assignProject", assign.Project),
					kv("assignOption", assign.Option),
				)
			}
		}
	}

	option := &projectOption{
		Items:     items,
		evaluator: evaluator,
	}
	return option, nil
}

// endregion
