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
	Items []*profileOptionItemSetting
}

func newProfileOptionSetting(items []*profileOptionItemSetting) *profileOptionSetting {
	return &profileOptionSetting{
		Items: items,
	}
}

func (s *profileOptionSetting) merge(setting *profileOptionSetting) {
	s.Items = append(s.Items, setting.Items...)
}

func (s *profileOptionSetting) getItems(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if _, exist := items[item.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, errW(err, "get profile option specify items error",
				reason("eval expr error"),
				kv("item", item),
			)
		}
		if matched {
			items[item.Name] = item.Value
		}
	}
	return items, nil
}

func (s *profileOptionSetting) inspect() *ProfileOptionSettingInspection {
	var items []*ProfileOptionItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProfileOptionSettingInspection(items)
}

// endregion

// region profileOptionItemSetting

type profileOptionItemSetting struct {
	Name  string
	Value string
	Match string
	match *EvalExpr
}

func newProfileOptionItemSetting(name, value, match string, matchObj *EvalExpr) *profileOptionItemSetting {
	return &profileOptionItemSetting{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s *profileOptionItemSetting) inspect() *ProfileOptionItemSettingInspection {
	return newProfileOptionItemSettingInspection(s.Name, s.Value, s.Match)
}

// endregion

// region profileOptionSettingModel

type profileOptionSettingModel struct {
	Items []*profileOptionItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProfileOptionSettingModel(items []*profileOptionItemSettingModel) *profileOptionSettingModel {
	return &profileOptionSettingModel{
		Items: items,
	}
}

func (m *profileOptionSettingModel) convert(ctx *modelConvertContext) (*profileOptionSetting, error) {
	var items []*profileOptionItemSetting
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(ctx.ChildItem("items", i))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return newProfileOptionSetting(items), nil
}

// endregion

// region profileOptionItemSettingModel

type profileOptionItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func (m *profileOptionItemSettingModel) convert(ctx *modelConvertContext) (setting *profileOptionItemSetting, err error) {
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

	return newProfileOptionItemSetting(m.Name, m.Value, m.Match, matchObj), nil
}

// endregion
