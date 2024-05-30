package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

type appProfile struct {
	logger                          *Logger
	workspace                       *Workspace
	evaluator                       *Evaluator
	workspaceShellEntities          workspaceShellEntitySet
	workspaceImportRegistryEntities workspaceImportRegistryEntitySet
	workspaceImportRedirectEntities workspaceImportRedirectEntitySet
	projectOptionSpecifyEntities    projectOptionSpecifyEntitySet
	projectScriptSourceEntities     projectSourceEntitySet
	projectScriptImportEntities     projectImportEntitySet
	projectConfigSourceEntities     projectSourceEntitySet
	projectConfigImportEntities     projectImportEntitySet
	projectManifestsByPath          map[string]*ProjectManifest
	projectManifestsByName          map[string]*ProjectManifest
}

func newAppProfile(workspace *Workspace, manifests []*AppProfileManifest) *appProfile {
	workspaceShellEntities := workspaceShellEntitySet{}
	workspaceImportRegistryEntities := workspaceImportRegistryEntitySet{}
	workspaceImportRedirectEntities := workspaceImportRedirectEntitySet{}
	projectOptionSpecifyEntities := projectOptionSpecifyEntitySet{}
	projectScriptSourceEntities := projectSourceEntitySet{}
	projectScriptImportEntities := projectImportEntitySet{}
	projectConfigSourceEntities := projectSourceEntitySet{}
	projectConfigImportEntities := projectImportEntitySet{}
	for i := 0; i < len(manifests); i++ {
		manifest := manifests[i]
		workspaceShellEntities.merge(manifest.Workspace.Shell.entities)
		workspaceImportRegistryEntities.merge(manifest.Workspace.Import.registryEntities)
		workspaceImportRedirectEntities = append(workspaceImportRedirectEntities, manifest.Workspace.Import.redirectEntities...)
		projectOptionSpecifyEntities = append(projectOptionSpecifyEntities, manifest.Project.Option.entities...)
		projectScriptSourceEntities = append(projectScriptSourceEntities, manifest.Project.Script.sourceEntities...)
		projectScriptImportEntities = append(projectScriptImportEntities, manifest.Project.Script.importEntities...)
		projectConfigSourceEntities = append(projectConfigSourceEntities, manifest.Project.Config.sourceEntities...)
		projectConfigImportEntities = append(projectConfigImportEntities, manifest.Project.Config.importEntities...)
	}
	workspaceShellEntities.merge(workspace.manifest.Shell.entities)
	workspaceShellEntities.mergeDefault()
	workspaceImportRegistryEntities.merge(workspace.manifest.Import.registryEntities)
	workspaceImportRegistryEntities.mergeDefault()
	workspaceImportRedirectEntities = append(workspaceImportRedirectEntities, workspace.manifest.Import.redirectEntities...)

	profile := &appProfile{
		logger:                          workspace.logger,
		workspace:                       workspace,
		evaluator:                       workspace.evaluator,
		workspaceShellEntities:          workspaceShellEntities,
		workspaceImportRegistryEntities: workspaceImportRegistryEntities,
		workspaceImportRedirectEntities: workspaceImportRedirectEntities,
		projectOptionSpecifyEntities:    projectOptionSpecifyEntities,
		projectScriptSourceEntities:     projectScriptSourceEntities,
		projectScriptImportEntities:     projectScriptImportEntities,
		projectConfigSourceEntities:     projectConfigSourceEntities,
		projectConfigImportEntities:     projectConfigImportEntities,
		projectManifestsByPath:          map[string]*ProjectManifest{},
		projectManifestsByName:          map[string]*ProjectManifest{},
	}
	return profile
}

func (p *appProfile) makeAppOption(projectName string) (*appOption, error) {
	specifyItems, err := p.projectOptionSpecifyEntities.getItems(p.evaluator)
	if err != nil {
		return nil, err
	}
	option := newAppOption(p.workspace.systemInfo, p.evaluator, projectName, specifyItems)
	return option, nil
}

func (p *appProfile) getWorkspaceShellEntity(name string) (*workspaceShellEntity, error) {
	return p.workspaceShellEntities.getEntity(name, p.evaluator)
}

func (p *appProfile) getWorkspaceImportRegistryLink(registry *projectLinkRegistry) (*projectLink, error) {
	evaluator := p.evaluator.SetRootData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ref.Type,
		"refName": registry.ref.Name,
	})
	return p.workspaceImportRegistryEntities.getLink(registry.Name, evaluator)
}

func (p *appProfile) getWorkspaceImportRedirectLink(resources []string) (*projectLink, string, error) {
	return p.workspaceImportRedirectEntities.getLink(resources, p.evaluator)
}

func (p *appProfile) getProjectLinkTarget(link *projectLink) (target *projectLinkTarget, err error) {
	finalLink := link
	if link.Registry != nil {
		registryLink, err := p.getWorkspaceImportRegistryLink(link.Registry)
		if err != nil {
			return nil, err
		}
		if registryLink == nil {
			return nil, errN("resolve project link error",
				reason("registry not found"),
				kv("link", link),
			)
		}
		finalLink = registryLink
	}
	path := ""
	if finalLink.Dir != nil {
		path = finalLink.Dir.Path
	} else if finalLink.Git != nil {
		path = p.workspace.getGitProjectPath(finalLink.Git.parsedUrl, finalLink.Git.parsedRef)
	} else {
		impossible()
	}
	resources := []string{
		finalLink.Normalized,
		"dir:" + path,
	}
	redirectLink, _, err := p.getWorkspaceImportRedirectLink(resources)
	if err != nil {
		return nil, err
	}
	if redirectLink != nil {
		finalLink = redirectLink
		if finalLink.Dir != nil {
			path = finalLink.Dir.Path
		} else if finalLink.Git != nil {
			path = p.workspace.getGitProjectPath(finalLink.Git.parsedUrl, finalLink.Git.parsedRef)
		} else {
			impossible()
		}
	}
	target = &projectLinkTarget{
		Link: link,
		Path: path,
		Git:  finalLink.Git,
	}
	return target, nil
}

func (p *appProfile) getProjectManifestByRawLink(rawLink string) (manifest *ProjectManifest, err error) {
	link, err := parseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	target, err := p.getProjectLinkTarget(link)
	if err != nil {
		return nil, err
	}
	return p.getProjectManifestByLinkTarget(target)
}

func (p *appProfile) getProjectManifestByLinkTarget(target *projectLinkTarget) (manifest *ProjectManifest, err error) {
	if target.Git != nil {
		return p.getProjectManifestByGit(target.Path, target.Git.Url, target.Git.parsedUrl, target.Git.Ref, target.Git.parsedRef)
	} else {
		return p.getProjectManifestByDir(target.Path)
	}
}

func (p *appProfile) getProjectManifestByDir(path string) (manifest *ProjectManifest, err error) {
	if !dsh_utils.IsDirExists(path) {
		return nil, errN("load project manifest error",
			reason("project dir not exists"),
			kv("path", path),
		)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("get abs-path error"),
			kv("path", path),
		)
	}
	path = absPath
	if m, exist := p.projectManifestsByPath[path]; exist {
		return m, nil
	}
	p.logger.DebugDesc("load project manifest", kv("path", path))
	if manifest, err = loadProjectManifest(path); err != nil {
		return nil, err
	}
	if existManifest, exist := p.projectManifestsByName[manifest.projectName]; exist {
		if existManifest.projectPath != manifest.projectPath {
			return nil, errN("load project manifest error",
				reason("project name duplicated"),
				kv("projectName", manifest.projectName),
				kv("projectPath1", manifest.projectPath),
				kv("projectPath2", existManifest.projectPath),
			)
		}
	}
	p.projectManifestsByPath[manifest.projectPath] = manifest
	p.projectManifestsByName[manifest.projectName] = manifest
	return manifest, nil
}

func (p *appProfile) getProjectManifestByGit(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *projectLinkGitRef) (manifest *ProjectManifest, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, errW(err, "load project manifest error",
				reason("parse url error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
			)
		}
	}
	if parsedRef == nil {
		if parsedRef, err = parseProjectLinkGitRef(rawRef); err != nil {
			return nil, errW(err, "load project manifest error",
				reason("parse ref error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
			)
		}
	}
	if path == "" {
		path = p.workspace.getGitProjectPath(parsedUrl, parsedRef)
	}
	if err = p.workspace.downloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, errW(err, "load project manifest error",
			reason("download project error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	manifest, err = p.getProjectManifestByDir(path)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	return manifest, nil
}
