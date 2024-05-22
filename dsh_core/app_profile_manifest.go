package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"github.com/expr-lang/expr/vm"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
)

// region manifest

type AppProfileManifest struct {
	Option       *AppProfileManifestOption
	Shell        AppProfileManifestShell
	Import       *AppProfileManifestImport
	Script       *AppProfileManifestProjectScript
	Config       *AppProfileManifestProjectConfig
	manifestPath string
	manifestType manifestMetadataType
}

func loadAppProfileManifest(path string) (*AppProfileManifest, error) {
	manifest := &AppProfileManifest{
		Option: &AppProfileManifestOption{},
		Shell:  AppProfileManifestShell{},
		Import: &AppProfileManifestImport{},
		Script: &AppProfileManifestProjectScript{},
		Config: &AppProfileManifestProjectConfig{},
	}

	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, errW(err, "load app profile manifest error",
				reason("get abs-path error"),
				kv("path", path),
			)
		}
		path = absPath
	}

	if path != "" {
		metadata, err := loadManifestFromFile(path, "", manifest)
		if err != nil {
			return nil, errW(err, "load app profile manifest error",
				reason("load manifest from file error"),
				kv("path", path),
			)
		}
		manifest.manifestPath = metadata.ManifestPath
		manifest.manifestType = metadata.ManifestType
	}

	if err := manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func MakeAppProfileManifest(optionValues map[string]string, shell AppProfileManifestShell, importRegistries []*AppProfileManifestImportRegistry, importRedirects []*AppProfileManifestImportRedirect) (*AppProfileManifest, error) {
	manifest := &AppProfileManifest{
		Option: &AppProfileManifestOption{
			Values: optionValues,
		},
		Shell: shell,
		Import: &AppProfileManifestImport{
			Registries: importRegistries,
			Redirects:  importRedirects,
		},
		Script: &AppProfileManifestProjectScript{},
		Config: &AppProfileManifestProjectConfig{},
	}

	if err := manifest.init(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *AppProfileManifest) DescExtraKeyValues() KVS {
	return KVS{
		kv("manifestPath", m.manifestPath),
		kv("manifestType", m.manifestType),
	}
}

func (m *AppProfileManifest) init() (err error) {
	if err = m.Option.init(m); err != nil {
		return err
	}
	if err = m.Shell.init(m); err != nil {
		return err
	}
	if err = m.Import.init(m); err != nil {
		return err
	}
	if err = m.Script.init(m); err != nil {
		return err
	}
	if err = m.Config.init(m); err != nil {
		return err
	}

	return nil
}

// endregion

// region option

type AppProfileManifestOption struct {
	Values map[string]string
}

func (o *AppProfileManifestOption) init(manifest *AppProfileManifest) error {
	return nil
}

// endregion

// region shell

type AppProfileManifestShell map[string]*AppProfileManifestShellItem

type AppProfileManifestShellItem struct {
	Path string
	Exts []string
	Args []string
}

func (s AppProfileManifestShell) init(manifest *AppProfileManifest) (err error) {
	for k, v := range s {
		if v.Path != "" && !dsh_utils.IsFileExists(v.Path) {
			return errN("app profile manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("shell.%s.path", k)),
				kv("value", v.Path),
			)
		}
		for i := 0; i < len(v.Exts); i++ {
			if v.Exts[i] == "" {
				return errN("app profile manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("shell.%s.exts[%d]", k, i)),
				)
			}
		}
		for i := 0; i < len(v.Args); i++ {
			if v.Args[i] == "" {
				return errN("app profile manifest invalid",
					reason("value empty"),
					kv("path", manifest.manifestPath),
					kv("field", fmt.Sprintf("shell.%s.args[%d]", k, i)),
				)
			}
		}
	}

	return nil
}

func (s AppProfileManifestShell) getPath(shell string) string {
	if i, exist := s[shell]; exist {
		return i.Path
	}
	return ""
}

func (s AppProfileManifestShell) getExts(shell string) []string {
	if i, exist := s[shell]; exist {
		if i.Exts != nil {
			return i.Exts
		}
	}
	return nil
}

func (s AppProfileManifestShell) getArgs(shell string) []string {
	if i, ok := s[shell]; ok {
		if i.Args != nil {
			return i.Args
		}
	}
	return nil
}

// endregion

// region import

type AppProfileManifestImport struct {
	Registries          []*AppProfileManifestImportRegistry
	Redirects           []*AppProfileManifestImportRedirect
	registryDefinitions map[string]*importRegistryDefinition
	redirectDefinitions []*importRedirectDefinition
}

type AppProfileManifestImportRegistry struct {
	Name  string
	Local *AppProfileManifestImportLocal
	Git   *AppProfileManifestImportGit
}

type AppProfileManifestImportRedirect struct {
	Prefix string
	Local  *AppProfileManifestImportLocal
	Git    *AppProfileManifestImportGit
}

type AppProfileManifestImportLocal struct {
	Dir string
}

type AppProfileManifestImportGit struct {
	Url string
	Ref string
}

func (imp *AppProfileManifestImport) init(manifest *AppProfileManifest) (err error) {
	registryDefinitions := make(map[string]*importRegistryDefinition)
	for i := 0; i < len(imp.Registries); i++ {
		registry := imp.Registries[i]
		if registry.Name == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].name", i)),
			)
		}
		if _, exist := registryDefinitions[registry.Name]; exist {
			return errN("app profile manifest invalid",
				reason("value duplicate"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.registries[%d].name", i)),
				kv("value", registry.Name),
			)
		}
		if err = imp.checkImportMode(manifest, registry.Local, registry.Git, "registries", i); err != nil {
			return err
		}
		if registry.Git != nil {
			if registry.Git.Ref == "" {
				registry.Git.Ref = "main"
			}
		}
		if registry.Local != nil {
			registryDefinitions[registry.Name] = newImportRegistryLocalDefinition(registry.Name, registry.Local.Dir)
		} else if registry.Git != nil {
			registryDefinitions[registry.Name] = newImportRegistryGitDefinition(registry.Name, registry.Git.Url, registry.Git.Ref)
		} else {
			impossible()
		}
	}
	imp.registryDefinitions = registryDefinitions

	redirectPrefixes := make(map[string]bool)
	var redirectDefinitions []*importRedirectDefinition
	for i := 0; i < len(imp.Redirects); i++ {
		redirect := imp.Redirects[i]
		if redirect.Prefix == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].prefix", i)),
			)
		}
		if _, exist := redirectPrefixes[redirect.Prefix]; exist {
			return errN("app profile manifest invalid",
				reason("value duplicate"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.redirects[%d].prefix", i)),
				kv("value", redirect.Prefix),
			)
		}
		redirectPrefixes[redirect.Prefix] = true
		if err = imp.checkImportMode(manifest, redirect.Local, redirect.Git, "redirects", i); err != nil {
			return err
		}
		if redirect.Local != nil {
			redirectDefinitions = append(redirectDefinitions, newImportRedirectLocalDefinition(redirect.Prefix, redirect.Local.Dir))
		} else if redirect.Git != nil {
			redirectDefinitions = append(redirectDefinitions, newImportRedirectGitDefinition(redirect.Prefix, redirect.Git.Url, redirect.Git.Ref))
		} else {
			impossible()
		}
	}
	if len(redirectDefinitions) > 0 {
		slices.SortStableFunc(redirectDefinitions, func(l, r *importRedirectDefinition) int {
			return len(r.Prefix) - len(l.Prefix)
		})
		imp.redirectDefinitions = redirectDefinitions
	}

	return nil
}

func (imp *AppProfileManifestImport) checkImportMode(manifest *AppProfileManifest, local *AppProfileManifestImportLocal, git *AppProfileManifestImportGit, scope string, index int) error {
	importModeCount := 0
	if local != nil {
		importModeCount++
	}
	if git != nil {
		importModeCount++
	}
	if importModeCount != 1 {
		return errN("app profile manifest invalid",
			reason("[local, git] must have only one"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("import.%s[%d]", scope, index)),
		)
	} else if local != nil {
		if local.Dir == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.%s[%d].local.dir", scope, index)),
			)
		}
	} else {
		if git.Url == "" {
			return errN("app profile manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("import.%s[%d].git.url", scope, index)),
			)
		}
	}
	return nil
}

func (imp *AppProfileManifestImport) getRegistryDefinition(name string) *importRegistryDefinition {
	if definition, exist := imp.registryDefinitions[name]; exist {
		return definition
	}
	return nil
}

func (imp *AppProfileManifestImport) getRedirectDefinition(resources []string) (*importRedirectDefinition, string) {
	for i := 0; i < len(resources); i++ {
		resource := resources[i]
		for j := 0; j < len(imp.redirectDefinitions); j++ {
			definition := imp.redirectDefinitions[j]
			if path, found := strings.CutPrefix(resource, definition.Prefix); found {
				return definition, path
			}
		}
	}
	return nil, ""
}

// endregion

// region project script

type AppProfileManifestProjectScript struct {
	Sources           []*AppProfileManifestProjectSource
	Imports           []*AppProfileManifestProjectImport
	sourceDefinitions []*projectSourceDefinition
	importDefinitions []*projectImportDefinition
}

func (s *AppProfileManifestProjectScript) init(manifest *AppProfileManifest) (err error) {
	var sourceDefinitions []*projectSourceDefinition
	for i := 0; i < len(s.Sources); i++ {
		src := s.Sources[i]
		if err = src.init(manifest, "script", i); err != nil {
			return err
		}
		sourceDefinitions = append(sourceDefinitions, src.definition)
	}

	var importDefinitions []*projectImportDefinition
	for i := 0; i < len(s.Imports); i++ {
		imp := s.Imports[i]
		if err = imp.init(manifest, "script", i); err != nil {
			return err
		}
		importDefinitions = append(importDefinitions, imp.definition)
	}

	s.sourceDefinitions = sourceDefinitions
	s.importDefinitions = importDefinitions
	return nil
}

// endregion

// region project config

type AppProfileManifestProjectConfig struct {
	Sources           []*AppProfileManifestProjectSource
	Imports           []*AppProfileManifestProjectImport
	sourceDefinitions []*projectSourceDefinition
	importDefinitions []*projectImportDefinition
}

func (c *AppProfileManifestProjectConfig) init(manifest *AppProfileManifest) (err error) {
	var sourceDefinitions []*projectSourceDefinition
	for i := 0; i < len(c.Sources); i++ {
		src := c.Sources[i]
		if err = src.init(manifest, "config", i); err != nil {
			return err
		}
		sourceDefinitions = append(sourceDefinitions, src.definition)
	}

	var importDefinitions []*projectImportDefinition
	for i := 0; i < len(c.Imports); i++ {
		imp := c.Imports[i]
		if err = imp.init(manifest, "config", i); err != nil {
			return err
		}
		importDefinitions = append(importDefinitions, imp.definition)
	}

	c.sourceDefinitions = sourceDefinitions
	c.importDefinitions = importDefinitions
	return nil
}

// endregion

// region project source

type AppProfileManifestProjectSource struct {
	Dir        string
	Files      []string
	Match      string
	definition *projectSourceDefinition
}

func (s *AppProfileManifestProjectSource) init(manifest *AppProfileManifest, scope string, index int) (err error) {
	if s.Dir == "" {
		return errN("project manifest invalid",
			reason("value empty"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.sources[%d].dir", scope, index)),
		)
	}
	var matchExpr *vm.Program
	if s.Match != "" {
		matchExpr, err = dsh_utils.CompileExpr(s.Match)
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.sources[%d].match", scope, index)),
				kv("value", s.Match),
			)
		}
	}

	s.definition = newProjectSourceDefinition(s.Dir, s.Files, s.Match, matchExpr)
	return nil
}

// endregion

// region project import

type AppProfileManifestProjectImport struct {
	Registry   *AppProfileManifestProjectImportRegistry
	Local      *AppProfileManifestProjectImportLocal
	Git        *AppProfileManifestProjectImportGit
	Match      string
	definition *projectImportDefinition
}

type AppProfileManifestProjectImportRegistry struct {
	Name string
	Path string
	Ref  string
}

type AppProfileManifestProjectImportLocal struct {
	Dir string
}

type AppProfileManifestProjectImportGit struct {
	Url string
	Ref string
}

func (i *AppProfileManifestProjectImport) init(manifest *AppProfileManifest, scope string, index int) (err error) {
	importModeCount := 0
	if i.Registry != nil {
		importModeCount++
	}
	if i.Local != nil {
		importModeCount++
	}
	if i.Git != nil {
		importModeCount++
	}
	var registryDefinition *projectImportRegistryDefinition
	var localDefinition *projectImportLocalDefinition
	var gitDefinition *projectImportGitDefinition
	if importModeCount != 1 {
		return errN("project manifest invalid",
			reason("[registry, local, git] must have only one"),
			kv("path", manifest.manifestPath),
			kv("field", fmt.Sprintf("%s.imports[%d]", scope, index)),
		)
	} else if i.Registry != nil {
		if i.Registry.Name == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].registry.name", scope, index)),
			)
		}
		if i.Registry.Path == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].registry.path", scope, index)),
			)
		}
		registryDefinition = newProjectImportRegistryDefinition(i.Registry.Name, i.Registry.Path, i.Registry.Ref)
	} else if i.Local != nil {
		if i.Local.Dir == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].local.dir", scope, index)),
			)
		}
		localDefinition = newProjectImportLocalDefinition(i.Local.Dir)
	} else if i.Git != nil {
		if i.Git.Url == "" {
			return errN("project manifest invalid",
				reason("value empty"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].git.url", scope, index)),
			)
		}
		if i.Git.Ref == "" {
			i.Git.Ref = "main"
		}
		var parsedUrl *url.URL
		if parsedUrl, err = url.Parse(i.Git.Url); err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].git.url", scope, index)),
				kv("value", i.Git.Url),
			)
		}
		parsedRef := parseGitRef(i.Git.Ref)
		gitDefinition = newProjectImportGitDefinition(i.Git.Url, parsedUrl, i.Git.Ref, parsedRef)
	}
	var matchExpr *vm.Program
	if i.Match != "" {
		matchExpr, err = dsh_utils.CompileExpr(i.Match)
		if err != nil {
			return errW(err, "project manifest invalid",
				reason("value invalid"),
				kv("path", manifest.manifestPath),
				kv("field", fmt.Sprintf("%s.imports[%d].match", scope, index)),
				kv("value", i.Match),
			)
		}
	}

	i.definition = newProjectImportDefinition(registryDefinition, localDefinition, gitDefinition, i.Match, matchExpr)
	return nil
}

// endregion
