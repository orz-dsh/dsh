package dsh_core

import (
	"dsh/dsh_utils"
	"os/exec"
)

// region default

var workspaceExecutorSettingsDefault = workspaceExecutorSettingSet{
	"cmd": {{
		Name: "cmd",
		Exts: []string{".cmd", ".bat"},
		Args: []string{"/C", "{{.target.path}}"},
	}},
	"pwsh": {{
		Name: "pwsh",
		Exts: []string{".ps1"},
		Args: []string{"-NoProfile", "-File", "{{.target.path}}"},
	}},
	"powershell": {{
		Name: "powershell",
		Exts: []string{".ps1"},
		Args: []string{"-NoProfile", "-File", "{{.target.path}}"},
	}},
	"*": {{
		Name: "*",
		Exts: []string{".sh"},
		Args: []string{},
	}},
}

// endregion

// region workspaceExecutorSetting

type workspaceExecutorSetting struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
	match *EvalExpr
}

type workspaceExecutorSettingSet map[string][]*workspaceExecutorSetting

func newWorkspaceExecutorSetting(name string, path string, exts []string, args []string, match string, matchObj *EvalExpr) *workspaceExecutorSetting {
	return &workspaceExecutorSetting{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
		match: matchObj,
	}
}

func (s *workspaceExecutorSetting) getArgs(evaluator *Evaluator) ([]string, error) {
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

func (s *workspaceExecutorSetting) inspect() *WorkspaceExecutorSettingInspection {
	return newWorkspaceExecutorSettingInspection(s.Name, s.Path, s.Exts, s.Args, s.Match)
}

func (s workspaceExecutorSettingSet) merge(settings workspaceExecutorSettingSet) {
	for name, list := range settings {
		s[name] = append(s[name], list...)
	}
}

func (s workspaceExecutorSettingSet) mergeDefault() {
	s.merge(workspaceExecutorSettingsDefault)
}

func (s workspaceExecutorSettingSet) getSetting(name string, evaluator *Evaluator) (*workspaceExecutorSetting, error) {
	result := &workspaceExecutorSetting{Name: name}
	settings := s[name]
	if wildcardSettings, exist := s["*"]; exist {
		settings = append(settings, wildcardSettings...)
	}
	for i := 0; i < len(settings); i++ {
		setting := settings[i]
		matched, err := evaluator.EvalBoolExpr(setting.match)
		if err != nil {
			return nil, errW(err, "get workspace executor setting error",
				reason("eval expr error"),
				kv("setting", setting),
			)
		}
		if matched {
			if result.Path == "" && setting.Path != "" {
				result.Path = setting.Path
			}
			if result.Exts == nil && setting.Exts != nil {
				result.Exts = setting.Exts
			}
			if result.Args == nil && setting.Args != nil {
				result.Args = setting.Args
			}
		}
	}
	if result.Path == "" {
		path, err := exec.LookPath(result.Name)
		if err != nil {
			return nil, errW(err, "get workspace executor setting error",
				reason("look path error"),
				kv("result", result),
			)
		}
		result.Path = path
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

func (s workspaceExecutorSettingSet) inspect() []*WorkspaceExecutorSettingInspection {
	var inspections []*WorkspaceExecutorSettingInspection
	for _, list := range s {
		for i := 0; i < len(list); i++ {
			inspections = append(inspections, list[i].inspect())
		}
	}
	return inspections
}

// endregion

// region workspaceExecutorSettingModel

type workspaceExecutorSettingModel struct {
	Items []*workspaceExecutorItemSettingModel `yaml:"items" toml:"items" json:"items"`
}

func newWorkspaceExecutorSettingModel(items []*workspaceExecutorItemSettingModel) *workspaceExecutorSettingModel {
	return &workspaceExecutorSettingModel{
		Items: items,
	}
}

func (m *workspaceExecutorSettingModel) convert(ctx *modelConvertContext) (workspaceExecutorSettingSet, error) {
	settings := workspaceExecutorSettingSet{}
	for i := 0; i < len(m.Items); i++ {
		item := m.Items[i]
		if setting, err := item.convert(ctx.ChildItem("items", i)); err != nil {
			return nil, err
		} else {
			settings[item.Name] = append(settings[item.Name], setting)
		}
	}

	return settings, nil
}

// endregion

// region workspaceExecutorItemSettingModel

type workspaceExecutorItemSettingModel struct {
	Name  string   `yaml:"name" toml:"name" json:"name"`
	Path  string   `yaml:"path" toml:"path" json:"path"`
	Exts  []string `yaml:"exts" toml:"exts" json:"exts"`
	Args  []string `yaml:"args" toml:"args" json:"args"`
	Match string   `yaml:"match" toml:"match" json:"match"`
}

func newWorkspaceExecutorItemSettingModel(name, path string, exts, args []string, match string) *workspaceExecutorItemSettingModel {
	return &workspaceExecutorItemSettingModel{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

func (m *workspaceExecutorItemSettingModel) convert(ctx *modelConvertContext) (setting *workspaceExecutorSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}

	if m.Path != "" && !dsh_utils.IsFileExists(m.Path) {
		return nil, ctx.Child("path").NewValueInvalidError(m.Path)
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

	return newWorkspaceExecutorSetting(m.Name, m.Path, m.Exts, m.Args, m.Match, matchObj), nil
}

// endregion

// region WorkspaceExecutorSettingInspection

type WorkspaceExecutorSettingInspection struct {
	Name  string   `yaml:"name" toml:"name" json:"name"`
	Path  string   `yaml:"path,omitempty" toml:"path,omitempty" json:"path,omitempty"`
	Exts  []string `yaml:"exts,omitempty" toml:"exts,omitempty" json:"exts,omitempty"`
	Args  []string `yaml:"args,omitempty" toml:"args,omitempty" json:"args,omitempty"`
	Match string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceExecutorSettingInspection(name, path string, exts, args []string, match string) *WorkspaceExecutorSettingInspection {
	return &WorkspaceExecutorSettingInspection{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

// endregion
