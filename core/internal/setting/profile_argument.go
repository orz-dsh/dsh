package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region base

var profileArgumentNameCheckRegex = regexp.MustCompile("^_?[a-z][a-z0-9_]*[a-z0-9]$")

// endregion

// region ProfileArgumentSetting

type ProfileArgumentSetting struct {
	Items []*ProfileArgumentItemSetting
}

func NewProfileArgumentSetting(items []*ProfileArgumentItemSetting) *ProfileArgumentSetting {
	return &ProfileArgumentSetting{
		Items: items,
	}
}

func (s *ProfileArgumentSetting) Merge(other *ProfileArgumentSetting) {
	s.Items = append(s.Items, other.Items...)
}

func (s *ProfileArgumentSetting) GetArguments(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if _, exist := items[item.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(item.Match)
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

func (s *ProfileArgumentSetting) Inspect() *ProfileArgumentSettingInspection {
	var items []*ProfileArgumentItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewProfileArgumentSettingInspection(items)
}

// endregion

// region ProfileArgumentItemSetting

type ProfileArgumentItemSetting struct {
	Name  string
	Value string
	Match string
}

func NewProfileArgumentItemSetting(name, value, match string) *ProfileArgumentItemSetting {
	return &ProfileArgumentItemSetting{
		Name:  name,
		Value: value,
		Match: match,
	}
}

func (s *ProfileArgumentItemSetting) Inspect() *ProfileArgumentItemSettingInspection {
	return NewProfileArgumentItemSettingInspection(s.Name, s.Value, s.Match)
}

// endregion

// region ProfileArgumentSettingModel

type ProfileArgumentSettingModel struct {
	Items []*ProfileArgumentItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProfileArgumentSettingModel(items []*ProfileArgumentItemSettingModel) *ProfileArgumentSettingModel {
	return &ProfileArgumentSettingModel{
		Items: items,
	}
}

func (m *ProfileArgumentSettingModel) Convert(helper *ModelHelper) (*ProfileArgumentSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewProfileArgumentSetting(items), nil
}

// endregion

// region ProfileArgumentItemSettingModel

type ProfileArgumentItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewProfileArgumentItemSettingModel(name, value, match string) *ProfileArgumentItemSettingModel {
	return &ProfileArgumentItemSettingModel{
		Name:  name,
		Value: value,
		Match: match,
	}
}

func (m *ProfileArgumentItemSettingModel) Convert(helper *ModelHelper) (*ProfileArgumentItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !profileArgumentNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}

	return NewProfileArgumentItemSetting(m.Name, m.Value, m.Match), nil
}

// endregion
