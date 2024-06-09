package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

// region appProfile

type appProfile struct {
	logger                          *Logger
	workspace                       *Workspace
	ProfileOptionSettings           profileOptionSettingSet
	ProfileProjectSettings          profileProjectSettingSet
	WorkspaceExecutorSettings       workspaceExecutorSettingSet
	WorkspaceImportRegistrySettings workspaceImportRegistrySettingSet
	WorkspaceImportRedirectSettings workspaceImportRedirectSettingSet
	projectSettingsByPath           map[string]*projectSetting
	projectSettingsByName           map[string]*projectSetting
}

func newProfileInstance(workspace *Workspace, settings profileSettingSet) *appProfile {
	profileOptionSettings := profileOptionSettingSet{}
	profileProjectSettings := profileProjectSettingSet{}
	workspaceExecutorSettings := workspaceExecutorSettingSet{}
	workspaceImportRegistrySettings := workspaceImportRegistrySettingSet{}
	workspaceImportRedirectSettings := workspaceImportRedirectSettingSet{}
	for i := 0; i < len(settings); i++ {
		setting := settings[i]
		profileOptionSettings = append(profileOptionSettings, setting.optionSettings...)
		profileProjectSettings = append(profileProjectSettings, setting.projectSettings...)
		workspaceExecutorSettings.merge(setting.workspaceExecutorSettings)
		workspaceImportRegistrySettings.merge(setting.workspaceImportRegistrySettings)
		workspaceImportRedirectSettings = append(workspaceImportRedirectSettings, setting.workspaceImportRedirectSettings...)
	}
	workspaceExecutorSettings.merge(workspace.setting.ExecutorSettings)
	workspaceExecutorSettings.mergeDefault()
	workspaceImportRegistrySettings.merge(workspace.setting.ImportRegistrySettings)
	workspaceImportRegistrySettings.mergeDefault()
	workspaceImportRedirectSettings = append(workspaceImportRedirectSettings, workspace.setting.ImportRedirectSettings...)

	profile := &appProfile{
		logger:                          workspace.logger,
		workspace:                       workspace,
		ProfileOptionSettings:           profileOptionSettings,
		ProfileProjectSettings:          profileProjectSettings,
		WorkspaceExecutorSettings:       workspaceExecutorSettings,
		WorkspaceImportRegistrySettings: workspaceImportRegistrySettings,
		WorkspaceImportRedirectSettings: workspaceImportRedirectSettings,
		projectSettingsByPath:           map[string]*projectSetting{},
		projectSettingsByName:           map[string]*projectSetting{},
	}
	return profile
}

func (p *appProfile) getAppOption(entity *projectSetting, evaluator *Evaluator) (*appOption, error) {
	specifyItems, err := p.ProfileOptionSettings.getItems(evaluator)
	if err != nil {
		return nil, err
	}
	option := newAppOption(p.workspace.global.systemInfo, evaluator, entity.Name, specifyItems)
	return option, nil
}

func (p *appProfile) getExtraProjectSettings(evaluator *Evaluator) (projectSettingSet, error) {
	projectEntities, err := p.ProfileProjectSettings.getProjectSettings(evaluator)
	if err != nil {
		return nil, err
	}
	return projectEntities, nil
}

func (p *appProfile) getWorkspaceExecutorSetting(name string) (*workspaceExecutorSetting, error) {
	return p.WorkspaceExecutorSettings.getSetting(name, p.workspace.evaluator)
}

func (p *appProfile) getWorkspaceImportRegistryLink(registry *projectLinkRegistry) (*projectLink, error) {
	evaluator := p.workspace.evaluator.SetRootData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ref.Type,
		"refName": registry.ref.Name,
	})
	return p.WorkspaceImportRegistrySettings.getLink(registry.Name, evaluator)
}

func (p *appProfile) getWorkspaceImportRedirectLink(resources []string) (*projectLink, string, error) {
	return p.WorkspaceImportRedirectSettings.getLink(resources, p.workspace.evaluator)
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

func (p *appProfile) getProjectEntityByRawLink(rawLink string) (*projectSetting, error) {
	link, err := parseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	target, err := p.getProjectLinkTarget(link)
	if err != nil {
		return nil, err
	}
	return p.getProjectSettingByLinkTarget(target)
}

func (p *appProfile) getProjectSettingByLinkTarget(target *projectLinkTarget) (*projectSetting, error) {
	if target.Git != nil {
		return p.getProjectEntityByGit(target.Path, target.Git.Url, target.Git.parsedUrl, target.Git.Ref, target.Git.parsedRef)
	} else {
		return p.getProjectEntityByDir(target.Path)
	}
}

func (p *appProfile) getProjectEntityByDir(path string) (*projectSetting, error) {
	if !dsh_utils.IsDirExists(path) {
		return nil, errN("load project setting error",
			reason("project dir not exists"),
			kv("path", path),
		)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errW(err, "load project setting error",
			reason("get abs-path error"),
			kv("path", path),
		)
	}
	path = absPath
	if setting, exist := p.projectSettingsByPath[path]; exist {
		return setting, nil
	}

	p.logger.DebugDesc("load project setting", kv("path", path))
	var setting *projectSetting
	if setting, err = loadProjectSetting(path); err != nil {
		return nil, err
	}
	if existSetting, exist := p.projectSettingsByName[setting.Name]; exist {
		if existSetting.Path != setting.Path {
			return nil, errN("get project setting error",
				reason("project name duplicated"),
				kv("projectName", setting.Name),
				kv("projectPath1", setting.Path),
				kv("projectPath2", existSetting.Path),
			)
		}
	}
	p.projectSettingsByPath[setting.Path] = setting
	p.projectSettingsByName[setting.Name] = setting
	return setting, nil
}

func (p *appProfile) getProjectEntityByGit(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *projectLinkGitRef) (entity *projectSetting, err error) {
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

func (p *appProfile) inspect() *AppProfileInspection {
	return newAppProfileInspection(
		p.ProfileOptionSettings.inspect(),
		p.ProfileProjectSettings.inspect(),
		p.WorkspaceExecutorSettings.inspect(),
		p.WorkspaceImportRegistrySettings.inspect(),
		p.WorkspaceImportRedirectSettings.inspect(),
	)
}

// endregion

// region AppProfileInspection

type AppProfileInspection struct {
	Options   []*ProfileOptionSettingInspection     `yaml:"options,omitempty" toml:"options,omitempty" json:"options,omitempty"`
	Projects  []*ProfileProjectSettingInspection    `yaml:"projects,omitempty" toml:"projects,omitempty" json:"projects,omitempty"`
	Workspace *AppProfileWorkspaceSettingInspection `yaml:"workspace,omitempty" toml:"workspace,omitempty" json:"workspace,omitempty"`
}

type AppProfileWorkspaceSettingInspection struct {
	Executors        []*WorkspaceExecutorSettingInspection       `yaml:"executors,omitempty" toml:"executors,omitempty" json:"executors,omitempty"`
	ImportRegistries []*WorkspaceImportRegistrySettingInspection `yaml:"importRegistries,omitempty" toml:"importRegistries,omitempty" json:"importRegistries,omitempty"`
	ImportRedirects  []*WorkspaceImportRedirectSettingInspection `yaml:"importRedirects,omitempty" toml:"importRedirects,omitempty" json:"importRedirects,omitempty"`
}

func newAppProfileInspection(options []*ProfileOptionSettingInspection, projects []*ProfileProjectSettingInspection, executors []*WorkspaceExecutorSettingInspection, importRegistries []*WorkspaceImportRegistrySettingInspection, importRedirects []*WorkspaceImportRedirectSettingInspection) *AppProfileInspection {
	return &AppProfileInspection{
		Options:  options,
		Projects: projects,
		Workspace: &AppProfileWorkspaceSettingInspection{
			Executors:        executors,
			ImportRegistries: importRegistries,
			ImportRedirects:  importRedirects,
		},
	}
}

// endregion
