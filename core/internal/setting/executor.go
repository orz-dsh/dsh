package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"os/exec"
)

// region default

var executorSettingDefault = NewExecutorSetting([]*ExecutorItemSetting{
	{
		Name: "cmd",
		Exts: []string{".cmd", ".bat"},
		Args: []string{"/C", "{{.target_file}}"},
	},
	{
		Name: "pwsh",
		Exts: []string{".ps1"},
		Args: []string{"-NoProfile", "-File", "{{.target_file}}"},
	},
	{
		Name: "powershell",
		Exts: []string{".ps1"},
		Args: []string{"-NoProfile", "-File", "{{.target_file}}"},
	},
	{
		Name: "*",
		Exts: []string{".sh"},
		Args: []string{},
	},
})

// endregion

// region ExecutorSetting

type ExecutorSetting struct {
	Items       []*ExecutorItemSetting
	itemsByName map[string][]*ExecutorItemSetting
}

func NewExecutorSetting(items []*ExecutorItemSetting) *ExecutorSetting {
	itemsByName := map[string][]*ExecutorItemSetting{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		itemsByName[item.Name] = append(itemsByName[item.Name], item)
	}
	return &ExecutorSetting{
		Items:       items,
		itemsByName: itemsByName,
	}
}

func (s *ExecutorSetting) Inspect() *ExecutorSettingInspection {
	var items []*ExecutorItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return NewExecutorSettingInspection(items)
}

func (s *ExecutorSetting) Merge(other *ExecutorSetting) {
	for i := 0; i < len(other.Items); i++ {
		item := other.Items[i]
		s.Items = append(s.Items, item)
		s.itemsByName[item.Name] = append(s.itemsByName[item.Name], item)
	}
}

func (s *ExecutorSetting) MergeDefault() {
	s.Merge(executorSettingDefault)
}

func (s *ExecutorSetting) GetItem(name string, evaluator *Evaluator) (*ExecutorItemSetting, error) {
	result := &ExecutorItemSetting{Name: name}
	items := s.itemsByName[name]
	if wildcardSettings, exist := s.itemsByName["*"]; exist {
		items = append(items, wildcardSettings...)
	}
	for i := 0; i < len(items); i++ {
		item := items[i]
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, ErrW(err, "get workspace executor item error",
				Reason("eval expr error"),
				KV("item", item),
			)
		}
		if matched {
			if result.File == "" && item.File != "" {
				result.File = item.File
			}
			if result.Exts == nil && item.Exts != nil {
				result.Exts = item.Exts
			}
			if result.Args == nil && item.Args != nil {
				result.Args = item.Args
			}
		}
	}
	if result.File == "" {
		path, err := exec.LookPath(result.Name)
		if err != nil {
			return nil, ErrW(err, "get workspace executor setting error",
				Reason("look path error"),
				KV("result", result),
			)
		}
		result.File = path
	}
	if result.Exts == nil {
		return nil, ErrN("get workspace executor setting error",
			Reason("exts not set"),
			KV("result", result),
		)
	}
	if result.Args == nil {
		return nil, ErrN("get workspace executor setting error",
			Reason("args not set"),
			KV("result", result),
		)
	}
	return result, nil
}

// endregion

// region ExecutorItemSetting

type ExecutorItemSetting struct {
	Name  string
	File  string
	Exts  []string
	Args  []string
	Match string
	match *EvalExpr
}

func NewExecutorItemSetting(name, file string, exts, args []string, match string, matchObj *EvalExpr) *ExecutorItemSetting {
	return &ExecutorItemSetting{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
		match: matchObj,
	}
}

func (s *ExecutorItemSetting) GetArgs(evaluator *Evaluator) ([]string, error) {
	var args []string
	for i := 0; i < len(s.Args); i++ {
		rawArg := s.Args[i]
		arg, err := evaluator.EvalStringTemplate(rawArg)
		if err != nil {
			return nil, ErrW(err, "get workspace executor setting args error",
				Reason("eval template error"),
				KV("args", s.Args),
				KV("index", i),
			)
		}
		args = append(args, arg)
	}
	return args, nil
}

func (s *ExecutorItemSetting) inspect() *ExecutorItemSettingInspection {
	return NewExecutorItemSettingInspection(s.Name, s.File, s.Exts, s.Args, s.Match)
}

// endregion

// region ExecutorSettingModel

type ExecutorSettingModel struct {
	Items []*ExecutorItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewExecutorSettingModel(items []*ExecutorItemSettingModel) *ExecutorSettingModel {
	return &ExecutorSettingModel{
		Items: items,
	}
}

func (m *ExecutorSettingModel) Convert(helper *ModelHelper) (*ExecutorSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewExecutorSetting(items), nil
}

// endregion

// region ExecutorItemSettingModel

type ExecutorItemSettingModel struct {
	Name  string   `yaml:"name" toml:"name" json:"name"`
	File  string   `yaml:"file,omitempty" toml:"file,omitempty" json:"file,omitempty"`
	Exts  []string `yaml:"exts,omitempty" toml:"exts,omitempty" json:"exts,omitempty"`
	Args  []string `yaml:"args,omitempty" toml:"args,omitempty" json:"args,omitempty"`
	Match string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewExecutorItemSettingModel(name, file string, exts, args []string, match string) *ExecutorItemSettingModel {
	return &ExecutorItemSettingModel{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

func (m *ExecutorItemSettingModel) Convert(helper *ModelHelper) (*ExecutorItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}

	if m.File != "" && !IsFileExists(m.File) {
		return nil, helper.Child("file").NewValueInvalidError(m.File)
	}

	if err := helper.CheckStringItemEmpty("exts", m.Exts); err != nil {
		return nil, err
	}

	if err := helper.CheckStringItemEmpty("args", m.Args); err != nil {
		return nil, err
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return NewExecutorItemSetting(m.Name, m.File, m.Exts, m.Args, m.Match, matchObj), nil
}

// endregion
