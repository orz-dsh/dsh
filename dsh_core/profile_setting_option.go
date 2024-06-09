package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var profileOptionNameCheckRegex = regexp.MustCompile("^_?[a-z][a-z0-9_]*[a-z]$")

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

func (s *profileOptionSetting) inspect() *ProfileOptionSettingInspection {
	return newProfileOptionInspection(s.Name, s.Value, s.Match)
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

func (s profileOptionSettingSet) inspect() []*ProfileOptionSettingInspection {
	var inspections []*ProfileOptionSettingInspection
	for i := 0; i < len(s); i++ {
		inspections = append(inspections, s[i].inspect())
	}
	return inspections
}

// endregion

// region profileOptionSettingModel

type profileOptionSettingModel struct {
	Items []*profileOptionItemSettingModel
}

func newProfileOptionSettingModel(items []*profileOptionItemSettingModel) *profileOptionSettingModel {
	return &profileOptionSettingModel{
		Items: items,
	}
}

func (m *profileOptionSettingModel) convert(ctx *modelConvertContext) (profileOptionSettingSet, error) {
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

func (m *profileOptionItemSettingModel) convert(ctx *modelConvertContext) (setting *profileOptionSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !profileOptionNameCheckRegex.MatchString(m.Name) {
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

// region ProfileOptionSettingInspection

type ProfileOptionSettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProfileOptionInspection(name, value, match string) *ProfileOptionSettingInspection {
	return &ProfileOptionSettingInspection{
		Name:  name,
		Value: value,
		Match: match,
	}
}

// endregion
