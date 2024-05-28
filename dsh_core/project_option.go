package dsh_core

type projectOption struct {
	Items     map[string]any
	evaluator *Evaluator
}

func makeProjectOption(context *appContext, manifest *projectManifest) (*projectOption, error) {
	items := context.Option.GenericItems.copy()
	for i := 0; i < len(manifest.Option.Items); i++ {
		item := manifest.Option.Items[i]
		result, err := context.Option.findResult(manifest, item)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("find option result error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("optionName", item.Name),
			)
		}
		items[item.Name] = result.ParsedValue
	}

	evaluator := context.evaluator.SetRootData("options", items)
	verifies := manifest.Option.verifies
	for i := 0; i < len(verifies); i++ {
		verify := verifies[i]
		result, err := evaluator.EvalBoolExpr(verify)
		if err != nil {
			return nil, errW(err, "load project options error",
				reason("eval verify error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("verify", verify.Source().Content()),
			)
		}
		if !result {
			return nil, errN("load project options error",
				reason("verify options error"),
				kv("projectName", manifest.Name),
				kv("projectPath", manifest.projectPath),
				kv("verify", verify.Source().Content()),
			)
		}
	}

	for i := 0; i < len(manifest.Option.Items); i++ {
		item := manifest.Option.Items[i]
		for j := 0; j < len(item.Assigns); j++ {
			assign := item.Assigns[j]
			if err := context.Option.addAssign(manifest.Name, item.Name, assign.Project, assign.Option, assign.mapping); err != nil {
				return nil, errW(err, "load project options error",
					reason("add option assign error"),
					kv("projectName", manifest.Name),
					kv("projectPath", manifest.projectPath),
					kv("optionName", item.Name),
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
