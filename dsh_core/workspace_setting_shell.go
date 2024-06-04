package dsh_core

import (
	"dsh/dsh_utils"
	"os/exec"
)

// region default

var workspaceShellSettingsDefault = workspaceShellSettingSet{
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
		Exts: []string{".sh"},
		Args: []string{},
	}},
}

// endregion

// region workspaceShellSetting

type workspaceShellSetting struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
	match *EvalExpr
}

type workspaceShellSettingSet map[string][]*workspaceShellSetting

func newWorkspaceShellSetting(name string, path string, exts []string, args []string, match string, matchObj *EvalExpr) *workspaceShellSetting {
	return &workspaceShellSetting{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
		match: matchObj,
	}
}

func (s *workspaceShellSetting) getArgs(evaluator *Evaluator) ([]string, error) {
	var args []string
	for i := 0; i < len(s.Args); i++ {
		rawArg := s.Args[i]
		arg, err := evaluator.EvalStringTemplate(rawArg)
		if err != nil {
			return nil, errW(err, "get workspace shell setting args error",
				reason("eval template error"),
				kv("args", s.Args),
				kv("index", i),
			)
		}
		args = append(args, arg)
	}
	return args, nil
}

func (s workspaceShellSettingSet) merge(settings workspaceShellSettingSet) {
	for name, list := range settings {
		s[name] = append(s[name], list...)
	}
}

func (s workspaceShellSettingSet) mergeDefault() {
	s.merge(workspaceShellSettingsDefault)
}

func (s workspaceShellSettingSet) getSetting(name string, evaluator *Evaluator) (*workspaceShellSetting, error) {
	result := &workspaceShellSetting{Name: name}
	settings := s[name]
	if wildcardSettings, exist := s["*"]; exist {
		settings = append(settings, wildcardSettings...)
	}
	for i := 0; i < len(settings); i++ {
		setting := settings[i]
		matched, err := evaluator.EvalBoolExpr(setting.match)
		if err != nil {
			return nil, errW(err, "get workspace shell setting error",
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
			return nil, errW(err, "get workspace shell setting error",
				reason("look path error"),
				kv("result", result),
			)
		}
		result.Path = path
	}
	if result.Exts == nil {
		return nil, errN("get workspace shell setting error",
			reason("exts not set"),
			kv("result", result),
		)
	}
	if result.Args == nil {
		return nil, errN("get workspace shell setting error",
			reason("args not set"),
			kv("result", result),
		)
	}
	return result, nil
}

// endregion

// region workspaceShellSettingModel

type workspaceShellSettingModel struct {
	Items []*workspaceShellItemSettingModel
}

func newWorkspaceShellSettingModel(items []*workspaceShellItemSettingModel) *workspaceShellSettingModel {
	return &workspaceShellSettingModel{
		Items: items,
	}
}

func (m *workspaceShellSettingModel) convert(ctx *ModelConvertContext) (workspaceShellSettingSet, error) {
	settings := workspaceShellSettingSet{}
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

// region workspaceShellItemSettingModel

type workspaceShellItemSettingModel struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
}

func newWorkspaceShellItemSettingModel(name, path string, exts, args []string, match string) *workspaceShellItemSettingModel {
	return &workspaceShellItemSettingModel{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

func (m *workspaceShellItemSettingModel) convert(ctx *ModelConvertContext) (setting *workspaceShellSetting, err error) {
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

	return newWorkspaceShellSetting(m.Name, m.Path, m.Exts, m.Args, m.Match, matchObj), nil
}

// endregion
