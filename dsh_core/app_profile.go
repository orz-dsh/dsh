package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

type appProfile struct {
	logger                          *Logger
	workspace                       *Workspace
	profileOptionSpecifyEntities    profileOptionSchemaSet
	profileProjectEntities          profileProjectSchemaSet
	workspaceShellEntities          workspaceShellSettingSet
	workspaceImportRegistryEntities workspaceImportRegistrySettingSet
	workspaceImportRedirectEntities workspaceImportRedirectSettingSet
	projectEntitiesByPath           map[string]*projectSchema
	projectEntitiesByName           map[string]*projectSchema
}

func newAppProfile(workspace *Workspace, manifests []*ProfilePref) *appProfile {
	profileOptionSpecifyEntities := profileOptionSchemaSet{}
	profileProjectEntities := profileProjectSchemaSet{}
	workspaceShellEntities := workspaceShellSettingSet{}
	workspaceImportRegistryEntities := workspaceImportRegistrySettingSet{}
	workspaceImportRedirectEntities := workspaceImportRedirectSettingSet{}
	for i := 0; i < len(manifests); i++ {
		manifest := manifests[i]
		profileOptionSpecifyEntities = append(profileOptionSpecifyEntities, manifest.Option.schemas...)
		profileProjectEntities = append(profileProjectEntities, manifest.Project.schemas...)
		workspaceShellEntities.merge(manifest.Workspace.Shell.schemas)
		workspaceImportRegistryEntities.merge(manifest.Workspace.Import.registrySchemas)
		workspaceImportRedirectEntities = append(workspaceImportRedirectEntities, manifest.Workspace.Import.redirectSchemas...)
	}
	workspaceShellEntities.merge(workspace.setting.Shell)
	workspaceShellEntities.mergeDefault()
	workspaceImportRegistryEntities.merge(workspace.setting.ImportRegistries)
	workspaceImportRegistryEntities.mergeDefault()
	workspaceImportRedirectEntities = append(workspaceImportRedirectEntities, workspace.setting.ImportRedirects...)

	profile := &appProfile{
		logger:                          workspace.logger,
		workspace:                       workspace,
		profileOptionSpecifyEntities:    profileOptionSpecifyEntities,
		profileProjectEntities:          profileProjectEntities,
		workspaceShellEntities:          workspaceShellEntities,
		workspaceImportRegistryEntities: workspaceImportRegistryEntities,
		workspaceImportRedirectEntities: workspaceImportRedirectEntities,
		projectEntitiesByPath:           map[string]*projectSchema{},
		projectEntitiesByName:           map[string]*projectSchema{},
	}
	return profile
}

func (p *appProfile) getAppOption(entity *projectSchema, evaluator *Evaluator) (*appOption, error) {
	specifyItems, err := p.profileOptionSpecifyEntities.getItems(evaluator)
	if err != nil {
		return nil, err
	}
	option := newAppOption(p.workspace.global.systemInfo, evaluator, entity.Name, specifyItems)
	return option, nil
}

func (p *appProfile) getExtraProjectEntities(evaluator *Evaluator) (projectSchemaSet, error) {
	projectEntities, err := p.profileProjectEntities.getProjectEntities(evaluator)
	if err != nil {
		return nil, err
	}
	return projectEntities, nil
}

func (p *appProfile) getWorkspaceShellEntity(name string) (*workspaceShellSetting, error) {
	return p.workspaceShellEntities.getSetting(name, p.workspace.evaluator)
}

func (p *appProfile) getWorkspaceImportRegistryLink(registry *projectLinkRegistry) (*projectLink, error) {
	evaluator := p.workspace.evaluator.SetRootData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ref.Type,
		"refName": registry.ref.Name,
	})
	return p.workspaceImportRegistryEntities.getLink(registry.Name, evaluator)
}

func (p *appProfile) getWorkspaceImportRedirectLink(resources []string) (*projectLink, string, error) {
	return p.workspaceImportRedirectEntities.getLink(resources, p.workspace.evaluator)
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
		path = p.workspace.getGitProjectDir(finalLink.Git.parsedUrl, finalLink.Git.parsedRef)
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
			path = p.workspace.getGitProjectDir(finalLink.Git.parsedUrl, finalLink.Git.parsedRef)
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

func (p *appProfile) getProjectEntityByRawLink(rawLink string) (*projectSchema, error) {
	link, err := parseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	target, err := p.getProjectLinkTarget(link)
	if err != nil {
		return nil, err
	}
	return p.getProjectEntityByLinkTarget(target)
}

func (p *appProfile) getProjectEntityByLinkTarget(target *projectLinkTarget) (*projectSchema, error) {
	if target.Git != nil {
		return p.getProjectEntityByGit(target.Path, target.Git.Url, target.Git.parsedUrl, target.Git.Ref, target.Git.parsedRef)
	} else {
		return p.getProjectEntityByDir(target.Path)
	}
}

func (p *appProfile) getProjectEntityByDir(path string) (*projectSchema, error) {
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
	if entity, exist := p.projectEntitiesByPath[path]; exist {
		return entity, nil
	}

	p.logger.DebugDesc("load project manifest", kv("path", path))
	var manifest *projectManifest
	if manifest, err = loadProjectManifest(path); err != nil {
		return nil, err
	}
	entity := manifest.entity
	if existEntity, exist := p.projectEntitiesByName[entity.Name]; exist {
		if existEntity.Path != entity.Path {
			return nil, errN("get project entity error",
				reason("project name duplicated"),
				kv("projectName", entity.Name),
				kv("projectPath1", entity.Path),
				kv("projectPath2", existEntity.Path),
			)
		}
	}
	p.projectEntitiesByPath[entity.Path] = entity
	p.projectEntitiesByName[entity.Name] = entity
	return entity, nil
}

func (p *appProfile) getProjectEntityByGit(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *projectLinkGitRef) (entity *projectSchema, err error) {
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
		path = p.workspace.getGitProjectDir(parsedUrl, parsedRef)
	}
	if err = p.workspace.downloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, errW(err, "load project manifest error",
			reason("download project error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	entity, err = p.getProjectEntityByDir(path)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	return entity, nil
}
