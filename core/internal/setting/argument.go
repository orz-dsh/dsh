package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region base

var argumentNameCheckRegex = regexp.MustCompile("^_?[a-z][a-z0-9_]*[a-z]$")

// endregion

// region ArgumentSetting

type ArgumentSetting struct {
	Items []*ArgumentItemSetting
}

func NewArgumentSetting(items []*ArgumentItemSetting) *ArgumentSetting {
	return &ArgumentSetting{
		Items: items,
	}
}

func (s *ArgumentSetting) Merge(other *ArgumentSetting) {
	s.Items = append(s.Items, other.Items...)
}

func (s *ArgumentSetting) GetArguments(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if _, exist := items[item.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, ErrW(err, "get arguments error",
				Reason("eval expr error"),
				KV("item", item),
			)
		}
		if matched {
			items[item.Name] = item.Value
		}
	}
	return items, nil
}

func (s *ArgumentSetting) Inspect() *ArgumentSettingInspection {
	var items []*ArgumentItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewArgumentSettingInspection(items)
}

// endregion

// region ArgumentItemSetting

type ArgumentItemSetting struct {
	Name  string
	Value string
	Match string
	match *EvalExpr
}

func NewArgumentItemSetting(name, value, match string, matchObj *EvalExpr) *ArgumentItemSetting {
	return &ArgumentItemSetting{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s *ArgumentItemSetting) Inspect() *ArgumentItemSettingInspection {
	return NewArgumentItemSettingInspection(s.Name, s.Value, s.Match)
}

// endregion

// region ArgumentSettingModel

type ArgumentSettingModel struct {
	Items []*ArgumentItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewArgumentSettingModel(items []*ArgumentItemSettingModel) *ArgumentSettingModel {
	return &ArgumentSettingModel{
		Items: items,
	}
}

func (m *ArgumentSettingModel) Convert(helper *ModelHelper) (*ArgumentSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewArgumentSetting(items), nil
}

// endregion

// region ArgumentItemSettingModel

type ArgumentItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewArgumentItemSettingModel(name, value, match string) *ArgumentItemSettingModel {
	return &ArgumentItemSettingModel{
		Name:  name,
		Value: value,
		Match: match,
	}
}

func (m *ArgumentItemSettingModel) Convert(helper *ModelHelper) (*ArgumentItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !argumentNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return NewArgumentItemSetting(m.Name, m.Value, m.Match, matchObj), nil
}

// endregion
