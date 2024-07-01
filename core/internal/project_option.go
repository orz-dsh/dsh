package internal

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
)

// region ProjectOption

type ProjectOption struct {
	Items     map[string]any
	evaluator *Evaluator
}

func NewProjectOption(core *ApplicationCore, setting *ProjectSetting) (*ProjectOption, error) {
	items := core.Option.Common.copy()
	for i := 0; i < len(setting.Option.Items); i++ {
		declare := setting.Option.Items[i]
		result, err := core.Option.findResult(setting.Name, declare)
		if err != nil {
			return nil, ErrW(err, "load project options error",
				Reason("find option result error"),
				KV("projectName", setting.Name),
				KV("projectPath", setting.Dir),
				KV("optionName", declare.Name),
			)
		}
		items[declare.Name] = result.ParsedValue
	}

	evaluator := core.Evaluator.SetRootData("option", items)
	for i := 0; i < len(setting.Option.Checks); i++ {
		check := setting.Option.Checks[i]
		result, err := evaluator.EvalBoolExpr(check.ExprObj)
		if err != nil {
			return nil, ErrW(err, "load project options error",
				Reason("eval check error"),
				KV("projectName", setting.Name),
				KV("projectPath", setting.Dir),
				KV("check", check),
			)
		}
		if !result {
			return nil, ErrN("load project options error",
				Reason("check options error"),
				KV("projectName", setting.Name),
				KV("projectPath", setting.Dir),
				KV("check", check),
			)
		}
	}

	for i := 0; i < len(setting.Option.Items); i++ {
		optionSetting := setting.Option.Items[i]
		for j := 0; j < len(optionSetting.Assigns); j++ {
			assignSetting := optionSetting.Assigns[j]
			if err := core.Option.Assign.AddItem(setting.Name, optionSetting.Name, assignSetting); err != nil {
				return nil, ErrW(err, "load project options error",
					Reason("add option assign error"),
					KV("projectName", setting.Name),
					KV("projectPath", setting.Dir),
					KV("optionName", optionSetting.Name),
					KV("assignTarget", assignSetting.Target),
				)
			}
		}
	}

	option := &ProjectOption{
		Items:     items,
		evaluator: evaluator,
	}
	return option, nil
}

func (e *ProjectOption) Inspect() *ProjectOptionInspection {
	return NewProjectOptionInspection(e.Items)
}

// endregion
