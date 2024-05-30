package dsh_core

import (
	"dsh/dsh_utils"
	"os/exec"
	"path/filepath"
	"regexp"
)

// region profile

type workspaceProfileEntity struct {
	File     string
	Optional bool
	Match    string
	match    *EvalExpr
}

type workspaceProfileEntitySet []*workspaceProfileEntity

func newWorkspaceProfileEntity(file string, optional bool, match string, matchObj *EvalExpr) *workspaceProfileEntity {
	return &workspaceProfileEntity{
		File:     file,
		Optional: optional,
		Match:    match,
		match:    matchObj,
	}
}

func (s workspaceProfileEntitySet) getFiles(evaluator *Evaluator) ([]string, error) {
	var files []string
	for i := 0; i < len(s); i++ {
		entity := s[i]
		if matched, err := evaluator.EvalBoolExpr(entity.match); err != nil {
			return nil, errW(err, "get workspace profile files error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		} else if matched {
			rawFile, err := evaluator.EvalStringTemplate(entity.File)
			if err != nil {
				return nil, errW(err, "get workspace profile files error",
					reason("eval template error"),
					kv("entity", entity),
				)
			}
			file, err := filepath.Abs(rawFile)
			if err != nil {
				return nil, errW(err, "get workspace profile files error",
					reason("get abs-path error"),
					kv("entity", entity),
					kv("rawFile", rawFile),
				)
			}
			if dsh_utils.IsFileExists(file) {
				files = append(files, file)
			} else if !entity.Optional {
				return nil, errN("get workspace profile files error",
					reason("file not found"),
					kv("entity", entity),
					kv("rawFile", rawFile),
					kv("file", file),
				)
			}
		}
	}
	return files, nil
}

// endregion

// region shell

type workspaceShellEntity struct {
	Name  string
	Path  string
	Exts  []string
	Args  workspaceShellEntityArgs
	Match string
	match *EvalExpr
}

type workspaceShellEntityArgs []string

type workspaceShellEntitySet map[string][]*workspaceShellEntity

var workspaceShellEntitySetDefault = workspaceShellEntitySet{
	"cmd": {{
		Name: "cmd",
		Exts: []string{".cmd", ".bat"},
		Args: workspaceShellEntityArgs{"/C", "{{.target.path}}"},
	}},
	"pwsh": {{
		Name: "pwsh",
		Exts: []string{".ps1"},
		Args: workspaceShellEntityArgs{"-NoProfile", "-File", "{{.target.path}}"},
	}},
	"powershell": {{
		Name: "powershell",
		Exts: []string{".ps1"},
		Args: workspaceShellEntityArgs{"-NoProfile", "-File", "{{.target.path}}"},
	}},
	"*": {{
		Exts: []string{".sh"},
		Args: workspaceShellEntityArgs{},
	}},
}

func newWorkspaceShellEntity(name string, path string, exts []string, args []string, match string, matchObj *EvalExpr) *workspaceShellEntity {
	return &workspaceShellEntity{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
		match: matchObj,
	}
}

func (e *workspaceShellEntity) merge(entity *workspaceShellEntity) {
	if e.Path == "" && entity.Path != "" {
		e.Path = entity.Path
	}
	if e.Exts == nil && entity.Exts != nil {
		e.Exts = entity.Exts
	}
	if e.Args == nil && entity.Args != nil {
		e.Args = entity.Args
	}
}

func (s workspaceShellEntityArgs) getArgs(evaluator *Evaluator) ([]string, error) {
	var args []string
	for i := 0; i < len(s); i++ {
		rawArg := s[i]
		arg, err := evaluator.EvalStringTemplate(rawArg)
		if err != nil {
			return nil, errW(err, "get workspace shell entity args error",
				reason("eval template error"),
				kv("args", s),
				kv("index", i),
			)
		}
		args = append(args, arg)
	}
	return args, nil
}

func (s workspaceShellEntitySet) merge(entities workspaceShellEntitySet) {
	for name, list := range entities {
		s[name] = append(s[name], list...)
	}
}

func (s workspaceShellEntitySet) mergeDefault() {
	s.merge(workspaceShellEntitySetDefault)
}

func (s workspaceShellEntitySet) getEntity(name string, evaluator *Evaluator) (*workspaceShellEntity, error) {
	result := &workspaceShellEntity{Name: name}
	entities := s[name]
	if s, exist := s["*"]; exist {
		entities = append(entities, s...)
	}
	for i := 0; i < len(entities); i++ {
		entity := entities[i]
		matched, err := evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, errW(err, "get workspace shell entity error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		}
		if matched {
			result.merge(entity)
		}
	}
	if result.Path == "" {
		path, err := exec.LookPath(result.Name)
		if err != nil {
			return nil, errW(err, "get workspace shell entity error",
				reason("look path error"),
				kv("result", result),
			)
		}
		result.Path = path
	}
	if result.Exts == nil {
		return nil, errN("get workspace shell entity error",
			reason("exts not found"),
			kv("result", result),
		)
	}
	if result.Args == nil {
		return nil, errN("get workspace shell entity error",
			reason("args not found"),
			kv("result", result),
		)
	}
	return result, nil
}

// endregion

// region import

type workspaceImportRegistryEntity struct {
	Name  string
	Link  string
	Match string
	match *EvalExpr
}

type workspaceImportRegistryEntitySet map[string][]*workspaceImportRegistryEntity

type workspaceImportRedirectEntity struct {
	Regex string
	Link  string
	Match string
	regex *regexp.Regexp
	match *EvalExpr
}

type workspaceImportRedirectEntitySet []*workspaceImportRedirectEntity

var workspaceImportRegistryEntitySetDefault = workspaceImportRegistryEntitySet{
	"orz-dsh": {{
		Name: "orz-dsh",
		Link: "git:https://github.com/orz-dsh/{{.path}}.git#ref={{.ref}}",
	}},
	"orz-ops": {{
		Name: "orz-ops",
		Link: "git:https://github.com/orz-ops/{{.path}}.git#ref={{.ref}}",
	}},
}

func newWorkspaceImportRegistryEntity(name string, link string, match string, matchObj *EvalExpr) *workspaceImportRegistryEntity {
	return &workspaceImportRegistryEntity{
		Name:  name,
		Link:  link,
		Match: match,
		match: matchObj,
	}
}

func newWorkspaceImportRedirectEntity(regexStr string, link string, match string, regexObj *regexp.Regexp, matchObj *EvalExpr) *workspaceImportRedirectEntity {
	return &workspaceImportRedirectEntity{
		Regex: regexStr,
		Link:  link,
		Match: match,
		regex: regexObj,
		match: matchObj,
	}
}

func (s workspaceImportRegistryEntitySet) merge(entities workspaceImportRegistryEntitySet) {
	for name, list := range entities {
		s[name] = append(s[name], list...)
	}
}

func (s workspaceImportRegistryEntitySet) mergeDefault() {
	s.merge(workspaceImportRegistryEntitySetDefault)
}

func (s workspaceImportRegistryEntitySet) getLink(name string, evaluator *Evaluator) (*projectLink, error) {
	if entities, exist := s[name]; exist {
		for i := 0; i < len(entities); i++ {
			entity := entities[i]
			matched, err := evaluator.EvalBoolExpr(entity.match)
			if err != nil {
				return nil, errW(err, "get workspace import registry link error",
					reason("eval expr error"),
					kv("entity", entity),
				)
			}
			if matched {
				rawLink, err := evaluator.EvalStringTemplate(entity.Link)
				if err != nil {
					return nil, errW(err, "get workspace import registry link error",
						reason("eval template error"),
						kv("entity", entity),
					)
				}
				link, err := parseProjectLink(rawLink)
				if err != nil {
					return nil, errW(err, "get workspace import registry link error",
						reason("parse link error"),
						kv("entity", entity),
						kv("rawLink", rawLink),
					)
				}
				return link, nil
			}
		}
	}
	return nil, nil
}

func (s workspaceImportRedirectEntitySet) getLink(originals []string, evaluator *Evaluator) (*projectLink, string, error) {
	for i := 0; i < len(originals); i++ {
		original := originals[i]
		for j := 0; j < len(s); j++ {
			entity := s[j]
			matched, values := dsh_utils.RegexMatch(entity.regex, original)
			if !matched {
				continue
			}
			matched, err := evaluator.EvalBoolExpr(entity.match)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect link error",
					reason("eval expr error"),
					kv("entity", entity),
				)
			}
			if !matched {
				continue
			}
			evaluator2 := evaluator.SetData("regex", dsh_utils.MapStrStrToStrAny(values))
			rawLink, err := evaluator2.EvalStringTemplate(entity.Link)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect link error",
					reason("eval template error"),
					kv("entity", entity),
				)
			}
			link, err := parseProjectLink(rawLink)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect link error",
					reason("parse link error"),
					kv("entity", entity),
					kv("rawLink", rawLink),
				)
			}
			return link, original, nil
		}
	}
	return nil, "", nil
}

// endregion
