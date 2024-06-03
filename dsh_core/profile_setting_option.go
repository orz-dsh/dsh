package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var profileOptionNameCheckRegex = regexp.MustCompile("^_?[a-z][a-z0-9_]*$")

// endregion

// region profileOptionSetting

type profileOptionSetting struct {
	Name  string
	Value string
	Match string
	match *EvalExpr
}

type profileOptionSettingSet []*profileOptionSetting

func newProfileOptionSetting(name string, value string, match string, matchObj *EvalExpr) *profileOptionSetting {
	return &profileOptionSetting{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s profileOptionSettingSet) getItems(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s); i++ {
		entity := s[i]
		if _, exist := items[entity.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, errW(err, "get profile option specify items error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		}
		if matched {
			items[entity.Name] = entity.Value
		}
	}
	return items, nil
}

// endregion

// region profileOptionSettingModel

type profileOptionSettingModel struct {
	Items []*profileOptionItemSettingModel
}

func (m *profileOptionSettingModel) convert(ctx *ModelConvertContext) (profileOptionSettingSet, error) {
	settings := profileOptionSettingSet{}
	for i := 0; i < len(m.Items); i++ {
		if setting, err := m.Items[i].convert(ctx.ChildItem("items", i)); err != nil {
			return nil, err
		} else {
			settings = append(settings, setting)
		}
	}
	return settings, nil
}

// endregion

// region profileOptionItemSettingModel

type profileOptionItemSettingModel struct {
	Name  string
	Value string
	Match string
}

func (m *profileOptionItemSettingModel) convert(ctx *ModelConvertContext) (setting *profileOptionSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if checked := profileOptionNameCheckRegex.MatchString(m.Name); !checked {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newProfileOptionSetting(m.Name, m.Value, m.Match, matchObj), nil
}

// endregion
