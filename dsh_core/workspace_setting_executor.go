package dsh_core

import (
	"dsh/dsh_utils"
	"os/exec"
)

// region default

var workspaceExecutorSettingDefault = newWorkspaceExecutorSetting([]*workspaceExecutorItemSetting{
	{
		Name: "cmd",
		Exts: []string{".cmd", ".bat"},
		Args: []string{"/C", "{{.target.path}}"},
	},
	{
		Name: "pwsh",
		Exts: []string{".ps1"},
		Args: []string{"-NoProfile", "-File", "{{.target.path}}"},
	},
	{
		Name: "powershell",
		Exts: []string{".ps1"},
		Args: []string{"-NoProfile", "-File", "{{.target.path}}"},
	},
	{
		Name: "*",
		Exts: []string{".sh"},
		Args: []string{},
	},
})

// endregion

// region workspaceExecutorSetting

type workspaceExecutorSetting struct {
	Items       []*workspaceExecutorItemSetting
	itemsByName map[string][]*workspaceExecutorItemSetting
}

func newWorkspaceExecutorSetting(items []*workspaceExecutorItemSetting) *workspaceExecutorSetting {
	itemsByName := map[string][]*workspaceExecutorItemSetting{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		itemsByName[item.Name] = append(itemsByName[item.Name], item)
	}
	return &workspaceExecutorSetting{
		Items:       items,
		itemsByName: itemsByName,
	}
}

func (s *workspaceExecutorSetting) inspect() *WorkspaceExecutorSettingInspection {
	var items []*WorkspaceExecutorItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newWorkspaceExecutorSettingInspection(items)
}

func (s *workspaceExecutorSetting) merge(setting *workspaceExecutorSetting) {
	for i := 0; i < len(setting.Items); i++ {
		item := setting.Items[i]
		s.Items = append(s.Items, item)
		s.itemsByName[item.Name] = append(s.itemsByName[item.Name], item)
	}
}

func (s *workspaceExecutorSetting) mergeDefault() {
	s.merge(workspaceExecutorSettingDefault)
}

func (s *workspaceExecutorSetting) getItem(name string, evaluator *Evaluator) (*workspaceExecutorItemSetting, error) {
	result := &workspaceExecutorItemSetting{Name: name}
	items := s.itemsByName[name]
	if wildcardSettings, exist := s.itemsByName["*"]; exist {
		items = append(items, wildcardSettings...)
	}
	for i := 0; i < len(items); i++ {
		item := items[i]
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, errW(err, "get workspace executor item error",
				reason("eval expr error"),
				kv("item", item),
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
			return nil, errW(err, "get workspace executor setting error",
				reason("look path error"),
				kv("result", result),
			)
		}
		result.File = path
	}
	if result.Exts == nil {
		return nil, errN("get workspace executor setting error",
			reason("exts not set"),
			kv("result", result),
		)
	}
	if result.Args == nil {
		return nil, errN("get workspace executor setting error",
			reason("args not set"),
			kv("result", result),
		)
	}
	return result, nil
}

// endregion

// region workspaceExecutorItemSetting

type workspaceExecutorItemSetting struct {
	Name  string
	File  string
	Exts  []string
	Args  []string
	Match string
	match *EvalExpr
}

func newWorkspaceExecutorItemSetting(name, file string, exts, args []string, match string, matchObj *EvalExpr) *workspaceExecutorItemSetting {
	return &workspaceExecutorItemSetting{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
		match: matchObj,
	}
}

func (s *workspaceExecutorItemSetting) getArgs(evaluator *Evaluator) ([]string, error) {
	var args []string
	for i := 0; i < len(s.Args); i++ {
		rawArg := s.Args[i]
		arg, err := evaluator.EvalStringTemplate(rawArg)
		if err != nil {
			return nil, errW(err, "get workspace executor setting args error",
				reason("eval template error"),
				kv("args", s.Args),
				kv("index", i),
			)
		}
		args = append(args, arg)
	}
	return args, nil
}

func (s *workspaceExecutorItemSetting) inspect() *WorkspaceExecutorItemSettingInspection {
	return newWorkspaceExecutorItemSettingInspection(s.Name, s.File, s.Exts, s.Args, s.Match)
}

// endregion

// region workspaceExecutorSettingModel

type workspaceExecutorSettingModel struct {
	Items []*workspaceExecutorItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newWorkspaceExecutorSettingModel(items []*workspaceExecutorItemSettingModel) *workspaceExecutorSettingModel {
	return &workspaceExecutorSettingModel{
		Items: items,
	}
}

func (m *workspaceExecutorSettingModel) convert(ctx *modelConvertContext) (*workspaceExecutorSetting, error) {
	var items []*workspaceExecutorItemSetting
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(ctx.ChildItem("items", i))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return newWorkspaceExecutorSetting(items), nil
}

// endregion

// region workspaceExecutorItemSettingModel

type workspaceExecutorItemSettingModel struct {
	Name  string   `yaml:"name" toml:"name" json:"name"`
	File  string   `yaml:"file,omitempty" toml:"file,omitempty" json:"file,omitempty"`
	Exts  []string `yaml:"exts,omitempty" toml:"exts,omitempty" json:"exts,omitempty"`
	Args  []string `yaml:"args,omitempty" toml:"args,omitempty" json:"args,omitempty"`
	Match string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceExecutorItemSettingModel(name, file string, exts, args []string, match string) *workspaceExecutorItemSettingModel {
	return &workspaceExecutorItemSettingModel{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

func (m *workspaceExecutorItemSettingModel) convert(ctx *modelConvertContext) (_ *workspaceExecutorItemSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}

	if m.File != "" && !dsh_utils.IsFileExists(m.File) {
		return nil, ctx.Child("file").NewValueInvalidError(m.File)
	}

	for i := 0; i < len(m.Exts); i++ {
		if m.Exts[i] == "" {
			return nil, ctx.ChildItem("exts", i).NewValueEmptyError()
		}
	}

	for i := 0; i < len(m.Args); i++ {
		if m.Args[i] == "" {
			return nil, ctx.ChildItem("args", i).NewValueEmptyError()
		}
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newWorkspaceExecutorItemSetting(m.Name, m.File, m.Exts, m.Args, m.Match, matchObj), nil
}

// endregion
