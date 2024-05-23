package dsh_core

import (
	"github.com/expr-lang/expr/vm"
	"net/url"
	"os/exec"
	"strings"
)

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

func (d *workspaceShellDefinition) isCompleted() bool {
	return d.Path != "" && d.Exts != nil && d.Args != nil
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

func (ds workspaceShellDefinitions) fillDefinition(target *workspaceShellDefinition, matcher *Matcher) error {
	// TODO: priority
	if definitions, exist := ds[target.Name]; exist {
		for i := 0; i < len(definitions); i++ {
			definition := definitions[i]
			matched, err := matcher.Match(definition.match)
			if err != nil {
				return err
			}
			if matched {
				if definition.Path != "" {
					target.Path = definition.Path
				}
				if definition.Exts != nil {
					target.Exts = definition.Exts
				}
				if definition.Args != nil {
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
	Local *workspaceImportLocalDefinition
	Git   *workspaceImportGitDefinition
	Match string
	match *vm.Program
}

type workspaceImportRegistryDefinitions map[string][]*workspaceImportRegistryDefinition

type workspaceImportRedirectDefinition struct {
	Prefix string
	Local  *workspaceImportLocalDefinition
	Git    *workspaceImportGitDefinition
	Match  string
	match  *vm.Program
}

type workspaceImportRedirectDefinitions []*workspaceImportRedirectDefinition

type workspaceImportLocalDefinition struct {
	Dir string
}

type workspaceImportGitDefinition struct {
	Url string
	Ref string
}

var workspaceImportRegistryDefinitionsDefault = map[string]*workspaceImportRegistryDefinition{
	"orz-dsh": {
		Name: "orz-dsh",
		Git: &workspaceImportGitDefinition{
			Url: "https://github.com/orz-dsh/{{.path}}.git",
			Ref: "main",
		},
	},
	"orz-ops": {
		Name: "orz-ops",
		Git: &workspaceImportGitDefinition{
			Url: "https://github.com/orz-ops/{{.path}}.git",
			Ref: "main",
		},
	},
}

func newWorkspaceImportRegistryDefinition(name string, local *workspaceImportLocalDefinition, git *workspaceImportGitDefinition, match string, matchExpr *vm.Program) *workspaceImportRegistryDefinition {
	return &workspaceImportRegistryDefinition{
		Name:  name,
		Local: local,
		Git:   git,
		Match: match,
		match: matchExpr,
	}
}

func newWorkspaceImportRedirectDefinition(prefix string, local *workspaceImportLocalDefinition, git *workspaceImportGitDefinition, match string, matchExpr *vm.Program) *workspaceImportRedirectDefinition {
	return &workspaceImportRedirectDefinition{
		Prefix: prefix,
		Local:  local,
		Git:    git,
		Match:  match,
		match:  matchExpr,
	}
}

func newWorkspaceImportLocalDefinition(dir string) *workspaceImportLocalDefinition {
	return &workspaceImportLocalDefinition{
		Dir: dir,
	}
}

func newWorkspaceImportGitDefinition(url string, ref string) *workspaceImportGitDefinition {
	return &workspaceImportGitDefinition{
		Url: url,
		Ref: ref,
	}
}

func (ds workspaceImportRegistryDefinitions) getDefinition(name string, matcher *Matcher) (*workspaceImportRegistryDefinition, error) {
	// TODO: priority
	if definitions, exist := ds[name]; exist {
		for i := 0; i < len(definitions); i++ {
			definition := definitions[i]
			matched, err := matcher.Match(definition.match)
			if err != nil {
				return nil, err
			}
			if matched {
				return definition, nil
			}
		}
	}
	return nil, nil
}

func (ds workspaceImportRedirectDefinitions) getDefinition(resources []string, matcher *Matcher) (*workspaceImportRedirectDefinition, string, error) {
	// TODO: priority
	for i := 0; i < len(resources); i++ {
		resource := resources[i]
		for j := 0; j < len(ds); j++ {
			definition := ds[j]
			if path, found := strings.CutPrefix(resource, definition.Prefix); found {
				matched, err := matcher.Match(definition.match)
				if err != nil {
					return nil, "", err
				}
				if matched {
					return definition, path, nil
				}
			}
		}
	}
	return nil, "", nil
}

func getWorkspaceImportRegistryDefinitionDefault(name string) *workspaceImportRegistryDefinition {
	if definition, exist := workspaceImportRegistryDefinitionsDefault[name]; exist {
		return definition
	}
	return nil
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

func (ds projectOptionDefinitions) fillOptions(target map[string]string, matcher *Matcher) error {
	// TODO: priority
	for i := 0; i < len(ds); i++ {
		definition := ds[i]
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
	Registry *projectImportRegistryDefinition
	Local    *projectImportLocalDefinition
	Git      *projectImportGitDefinition
	Match    string
	match    *vm.Program
}

type projectImportRegistryDefinition struct {
	Name string
	Path string
	Ref  string
}

type projectImportLocalDefinition struct {
	Dir string
}

type projectImportGitDefinition struct {
	Url string
	Ref string
	url *url.URL
	ref *gitRef
}

func newProjectImportDefinition(registry *projectImportRegistryDefinition, local *projectImportLocalDefinition, git *projectImportGitDefinition, match string, matchExpr *vm.Program) *projectImportDefinition {
	return &projectImportDefinition{
		Registry: registry,
		Local:    local,
		Git:      git,
		Match:    match,
		match:    matchExpr,
	}
}

func newProjectImportRegistryDefinition(name string, path string, ref string) *projectImportRegistryDefinition {
	return &projectImportRegistryDefinition{
		Name: name,
		Path: path,
		Ref:  ref,
	}
}

func newProjectImportLocalDefinition(dir string) *projectImportLocalDefinition {
	return &projectImportLocalDefinition{
		Dir: dir,
	}
}

func newProjectImportGitDefinition(rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) *projectImportGitDefinition {
	return &projectImportGitDefinition{
		Url: rawUrl,
		Ref: rawRef,
		url: parsedUrl,
		ref: parsedRef,
	}
}

// endregion
