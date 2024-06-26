package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

// region appSetting

type appSetting struct {
	logger                *Logger
	workspace             *Workspace
	Option                *profileOptionSetting
	Project               *profileProjectSetting
	Executor              *workspaceExecutorSetting
	Registry              *workspaceRegistrySetting
	Redirect              *workspaceRedirectSetting
	projectSettingsByPath map[string]*projectSetting
	projectSettingsByName map[string]*projectSetting
}

func newAppSetting(workspace *Workspace, settings []*profileSetting) *appSetting {
	option := newProfileOptionSetting(nil)
	project := newProfileProjectSetting(nil)
	executor := newWorkspaceExecutorSetting(nil)
	registry := newWorkspaceRegistrySetting(nil)
	redirect := newWorkspaceRedirectSetting(nil)
	for i := 0; i < len(settings); i++ {
		setting := settings[i]
		option.merge(setting.Option)
		project.merge(setting.Project)
		executor.merge(setting.Executor)
		registry.merge(setting.Registry)
		redirect.merge(setting.Redirect)
	}
	executor.merge(workspace.setting.Executor)
	executor.mergeDefault()
	registry.merge(workspace.setting.Registry)
	registry.mergeDefault()
	redirect.merge(workspace.setting.Redirect)

	profile := &appSetting{
		logger:                workspace.logger,
		workspace:             workspace,
		Option:                option,
		Project:               project,
		Executor:              executor,
		Registry:              registry,
		Redirect:              redirect,
		projectSettingsByPath: map[string]*projectSetting{},
		projectSettingsByName: map[string]*projectSetting{},
	}
	return profile
}

func (s *appSetting) getAppOption(entity *projectSetting, evaluator *Evaluator) (*appOption, error) {
	specifyItems, err := s.Option.getItems(evaluator)
	if err != nil {
		return nil, err
	}
	option := newAppOption(s.workspace.global.systemInfo, evaluator, entity.Name, specifyItems)
	return option, nil
}

func (s *appSetting) getExtraProjectSettings(evaluator *Evaluator) ([]*projectSetting, error) {
	projectEntities, err := s.Project.getProjectSettings(evaluator)
	if err != nil {
		return nil, err
	}
	return projectEntities, nil
}

func (s *appSetting) getWorkspaceExecutorSetting(name string) (*workspaceExecutorItemSetting, error) {
	return s.Executor.getItem(name, s.workspace.evaluator)
}

func (s *appSetting) getWorkspaceImportRegistryLink(registry *projectLinkRegistry) (*projectLink, error) {
	evaluator := s.workspace.evaluator.SetRootData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ref.Type,
		"refName": registry.ref.Name,
	})
	return s.Registry.getLink(registry.Name, evaluator)
}

func (s *appSetting) getWorkspaceImportRedirectLink(resources []string) (*projectLink, string, error) {
	return s.Redirect.getLink(resources, s.workspace.evaluator)
}

func (s *appSetting) getProjectLinkTarget(link *projectLink) (target *projectLinkTarget, err error) {
	finalLink := link
	if link.Registry != nil {
		registryLink, err := s.getWorkspaceImportRegistryLink(link.Registry)
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
		path = s.workspace.getGitProjectDir(finalLink.Git.parsedUrl, finalLink.Git.parsedRef)
	} else {
		impossible()
	}
	resources := []string{
		finalLink.Normalized,
		"dir:" + path,
	}
	redirectLink, _, err := s.getWorkspaceImportRedirectLink(resources)
	if err != nil {
		return nil, err
	}
	if redirectLink != nil {
		finalLink = redirectLink
		if finalLink.Dir != nil {
			path = finalLink.Dir.Path
		} else if finalLink.Git != nil {
			path = s.workspace.getGitProjectDir(finalLink.Git.parsedUrl, finalLink.Git.parsedRef)
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

func (s *appSetting) getProjectEntityByRawLink(rawLink string) (*projectSetting, error) {
	link, err := parseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	target, err := s.getProjectLinkTarget(link)
	if err != nil {
		return nil, err
	}
	return s.getProjectSettingByLinkTarget(target)
}

func (s *appSetting) getProjectSettingByLinkTarget(target *projectLinkTarget) (*projectSetting, error) {
	if target.Git != nil {
		return s.getProjectEntityByGit(target.Path, target.Git.Url, target.Git.parsedUrl, target.Git.Ref, target.Git.parsedRef)
	} else {
		return s.getProjectEntityByDir(target.Path)
	}
}

func (s *appSetting) getProjectEntityByDir(path string) (*projectSetting, error) {
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
	if setting, exist := s.projectSettingsByPath[path]; exist {
		return setting, nil
	}

	s.logger.DebugDesc("load project setting", kv("path", path))
	var setting *projectSetting
	if setting, err = loadProjectSetting(path); err != nil {
		return nil, err
	}
	if existSetting, exist := s.projectSettingsByName[setting.Name]; exist {
		if existSetting.Dir != setting.Dir {
			return nil, errN("get project setting error",
				reason("project name duplicated"),
				kv("projectName", setting.Name),
				kv("projectPath1", setting.Dir),
				kv("projectPath2", existSetting.Dir),
			)
		}
	}
	s.projectSettingsByPath[setting.Dir] = setting
	s.projectSettingsByName[setting.Name] = setting
	return setting, nil
}

func (s *appSetting) getProjectEntityByGit(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *projectLinkGitRef) (entity *projectSetting, err error) {
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
		path = s.workspace.getGitProjectDir(parsedUrl, parsedRef)
	}
	if err = s.workspace.downloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, errW(err, "load project manifest error",
			reason("download project error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	entity, err = s.getProjectEntityByDir(path)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	return entity, nil
}

func (s *appSetting) inspect() *AppSettingInspection {
	return newAppSettingInspection(
		s.Option.inspect(),
		s.Project.inspect(),
		s.Executor.inspect(),
		s.Registry.inspect(),
		s.Redirect.inspect(),
	)
}

// endregion
