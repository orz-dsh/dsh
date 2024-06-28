package core

import "os/exec"

// region default

var executorSettingDefault = newExecutorSetting([]*executorItemSetting{
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

// region executorSetting

type executorSetting struct {
	Items       []*executorItemSetting
	itemsByName map[string][]*executorItemSetting
}

func newExecutorSetting(items []*executorItemSetting) *executorSetting {
	itemsByName := map[string][]*executorItemSetting{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		itemsByName[item.Name] = append(itemsByName[item.Name], item)
	}
	return &executorSetting{
		Items:       items,
		itemsByName: itemsByName,
	}
}

func (s *executorSetting) inspect() *WorkspaceExecutorSettingInspection {
	var items []*WorkspaceExecutorItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newWorkspaceExecutorSettingInspection(items)
}

func (s *executorSetting) merge(setting *executorSetting) {
	for i := 0; i < len(setting.Items); i++ {
		item := setting.Items[i]
		s.Items = append(s.Items, item)
		s.itemsByName[item.Name] = append(s.itemsByName[item.Name], item)
	}
}

func (s *executorSetting) mergeDefault() {
	s.merge(executorSettingDefault)
}

func (s *executorSetting) getItem(name string, evaluator *Evaluator) (*executorItemSetting, error) {
	result := &executorItemSetting{Name: name}
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

// region executorItemSetting

type executorItemSetting struct {
	Name  string
	File  string
	Exts  []string
	Args  []string
	Match string
	match *EvalExpr
}

func newExecutorItemSetting(name, file string, exts, args []string, match string, matchObj *EvalExpr) *executorItemSetting {
	return &executorItemSetting{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
		match: matchObj,
	}
}

func (s *executorItemSetting) getArgs(evaluator *Evaluator) ([]string, error) {
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

func (s *executorItemSetting) inspect() *WorkspaceExecutorItemSettingInspection {
	return newWorkspaceExecutorItemSettingInspection(s.Name, s.File, s.Exts, s.Args, s.Match)
}

// endregion
