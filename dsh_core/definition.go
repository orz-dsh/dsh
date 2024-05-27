package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/expr-lang/expr/vm"
	"os/exec"
	"path/filepath"
	"regexp"
)

// region workspace profile

type workspaceProfileDefinition struct {
	File     string
	Optional bool
	Match    string
	match    *vm.Program
}

type workspaceProfileDefinitions []*workspaceProfileDefinition

func newWorkspaceProfileDefinition(file string, optional bool, match string, matchExpr *vm.Program) *workspaceProfileDefinition {
	return &workspaceProfileDefinition{
		File:     file,
		Optional: optional,
		Match:    match,
		match:    matchExpr,
	}
}

func (ds workspaceProfileDefinitions) getFiles(matcher *dsh_utils.EvalMatcher, replacer *dsh_utils.EvalReplacer) ([]string, error) {
	// TODO: error
	var files []string
	for i := 0; i < len(ds); i++ {
		definition := ds[i]
		if matched, err := matcher.Match(definition.match); err != nil {
			return nil, err
		} else if matched {
			file, err := replacer.Replace(definition.File)
			if err != nil {
				return nil, err
			}
			absPath, err := filepath.Abs(file)
			if err != nil {
				return nil, err
			}
			if dsh_utils.IsFileExists(absPath) {
				files = append(files, absPath)
			} else if !definition.Optional {
				return nil, errN("workspace profile file not found",
					reason("file not found"),
					kv("definition", definition),
					kv("path", absPath),
				)
			}
		}
	}
	return files, nil
}

// endregion

// region workspace shell

type workspaceShellDefinition struct {
	Name  string
	Path  string
	Exts  []string
	Args  []string
	Match string
	match *vm.Program
}

type workspaceShellDefinitions map[string][]*workspaceShellDefinition

var workspaceShellExtsDefault = map[string][]string{
	"cmd":        {".cmd", ".bat"},
	"pwsh":       {".ps1"},
	"powershell": {".ps1"},
}

var workspaceShellExtsFallback = []string{".sh"}

var workspaceShellArgsDefault = map[string][]string{
	"cmd":        {"/C", "{{.target.path}}"},
	"pwsh":       {"-NoProfile", "-File", "{{.target.path}}"},
	"powershell": {"-NoProfile", "-File", "{{.target.path}}"},
}

func newWorkspaceShellDefinition(name string, path string, exts []string, args []string, match string, matchExpr *vm.Program) *workspaceShellDefinition {
	return &workspaceShellDefinition{
		Name:  name,
		Path:  path,
		Exts:  exts,
		Args:  args,
		Match: match,
		match: matchExpr,
	}
}

func newWorkspaceShellDefinitionEmpty(name string) *workspaceShellDefinition {
	return &workspaceShellDefinition{
		Name: name,
	}
}

func (d *workspaceShellDefinition) fillDefault() error {
	if d.Path == "" {
		path, err := exec.LookPath(d.Name)
		if err != nil {
			// TODO: error info
			return errW(err, "workspace shell definition fill default error",
				reason("look path error"),
				kv("name", d.Name),
			)
		}
		d.Path = path
	}
	if d.Exts == nil {
		if exts, exist := workspaceShellExtsDefault[d.Name]; exist {
			d.Exts = exts
		} else {
			d.Exts = workspaceShellExtsFallback
		}
	}
	if d.Args == nil {
		if args, exist := workspaceShellArgsDefault[d.Name]; exist {
			d.Args = args
		}
	}
	return nil
}

func (ds workspaceShellDefinitions) fillDefinition(target *workspaceShellDefinition, matcher *dsh_utils.EvalMatcher) error {
	// TODO: priority
	if definitions, exist := ds[target.Name]; exist {
		for i := 0; i < len(definitions); i++ {
			definition := definitions[i]
			matched, err := matcher.Match(definition.match)
			if err != nil {
				return err
			}
			if matched {
				// TODO: optimize
				// The priority of the previous one is higher than that of the later one.
				if target.Path == "" && definition.Path != "" {
					target.Path = definition.Path
				}
				if target.Exts == nil && definition.Exts != nil {
					target.Exts = definition.Exts
				}
				if target.Args == nil && definition.Args != nil {
					target.Args = definition.Args
				}
			}
		}
	}
	return nil
}

// endregion

// region workspace import

type workspaceImportRegistryDefinition struct {
	Name  string
	Link  string
	Match string
	match *vm.Program
}

type workspaceImportRegistryDefinitions map[string][]*workspaceImportRegistryDefinition

type workspaceImportRedirectDefinition struct {
	Regex *regexp.Regexp
	Link  string
	Match string
	match *vm.Program
}

type workspaceImportRedirectDefinitions []*workspaceImportRedirectDefinition

var workspaceImportRegistryDefinitionsDefault = workspaceImportRegistryDefinitions{
	"orz-dsh": {{
		Name: "orz-dsh",
		Link: "git:https://github.com/orz-dsh/{{.path}}.git#ref={{.ref}}",
	}},
	"orz-ops": {{
		Name: "orz-ops",
		Link: "git:https://github.com/orz-ops/{{.path}}.git#ref={{.ref}}",
	}},
}

func newWorkspaceImportRegistryDefinition(name string, link string, match string, matchExpr *vm.Program) *workspaceImportRegistryDefinition {
	return &workspaceImportRegistryDefinition{
		Name:  name,
		Link:  link,
		Match: match,
		match: matchExpr,
	}
}

func newWorkspaceImportRedirectDefinition(regex *regexp.Regexp, link string, match string, matchExpr *vm.Program) *workspaceImportRedirectDefinition {
	return &workspaceImportRedirectDefinition{
		Regex: regex,
		Link:  link,
		Match: match,
		match: matchExpr,
	}
}

func (ds workspaceImportRegistryDefinitions) getLink(name string, matcher *dsh_utils.EvalMatcher, replacer *dsh_utils.EvalReplacer) (*ProjectLink, error) {
	if definitions, exist := ds[name]; exist {
		for i := 0; i < len(definitions); i++ {
			definition := definitions[i]
			matched, err := matcher.Match(definition.match)
			if err != nil {
				return nil, err
			}
			if matched {
				rawLink, err := replacer.Replace(definition.Link)
				if err != nil {
					return nil, err
				}
				link, err := ParseProjectLink(rawLink)
				if err != nil {
					return nil, err
				}
				return link, nil
			}
		}
	}
	return nil, nil
}

func (ds workspaceImportRedirectDefinitions) getLink(links []string, matcher *dsh_utils.EvalMatcher, replacer *dsh_utils.EvalReplacer) (*ProjectLink, string, error) {
	// TODO: priority
	for i := 0; i < len(links); i++ {
		link := links[i]
		for j := 0; j < len(ds); j++ {
			definition := ds[j]
			matched, values := dsh_utils.RegexMatch(definition.Regex, link)
			if !matched {
				continue
			}
			matched, err := matcher.Match(definition.match)
			if err != nil {
				return nil, "", err
			}
			if matched {
				rawLink, err := replacer.ModifyData(func(data *dsh_utils.EvalData) *dsh_utils.EvalData {
					return data.Data("regex", dsh_utils.MapStrStrToMapStrAny(values))
				}).Replace(definition.Link)
				if err != nil {
					return nil, "", err
				}
				redirectLink, err := ParseProjectLink(rawLink)
				if err != nil {
					return nil, "", err
				}
				return redirectLink, link, nil
			}
		}
	}
	return nil, "", nil
}

// endregion

// region project option

type projectOptionDefinition struct {
	Name  string
	Value string
	Match string
	match *vm.Program
}

type projectOptionDefinitions []*projectOptionDefinition

func newProjectOptionDefinition(name string, value string, match string, matchExpr *vm.Program) *projectOptionDefinition {
	return &projectOptionDefinition{
		Name:  name,
		Value: value,
		Match: match,
		match: matchExpr,
	}
}

func (ds projectOptionDefinitions) fillOptions(target map[string]string, matcher *dsh_utils.EvalMatcher) error {
	for i := 0; i < len(ds); i++ {
		definition := ds[i]
		if _, exist := target[definition.Name]; exist {
			// The priority of the previous one is higher than that of the later one.
			continue
		}
		matched, err := matcher.Match(definition.match)
		if err != nil {
			return err
		}
		if matched {
			target[definition.Name] = definition.Value
		}
	}
	return nil
}

// endregion

// region project source

type projectSourceDefinition struct {
	Dir   string
	Files []string
	Match string
	match *vm.Program
}

func newProjectSourceDefinition(dir string, files []string, match string, matchExpr *vm.Program) *projectSourceDefinition {
	return &projectSourceDefinition{
		Dir:   dir,
		Files: files,
		Match: match,
		match: matchExpr,
	}
}

// endregion

// region project import

type projectImportDefinition struct {
	Link  *ProjectLink
	Match string
	match *vm.Program
}

func newProjectImportDefinition(link *ProjectLink, match string, matchExpr *vm.Program) *projectImportDefinition {
	return &projectImportDefinition{
		Link:  link,
		Match: match,
		match: matchExpr,
	}
}

// endregion
