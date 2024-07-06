package internal

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"net/url"
	"path/filepath"
)

// region ApplicationSetting

type ApplicationSetting struct {
	Logger         *Logger
	Workspace      *WorkspaceCore
	Argument       *ArgumentSetting
	Addition       *AdditionSetting
	Executor       *ExecutorSetting
	Registry       *RegistrySetting
	Redirect       *RedirectSetting
	projectsByPath map[string]*ProjectSetting
	projectsByName map[string]*ProjectSetting
}

func NewApplicationSetting(workspace *WorkspaceCore, profiles []*ProfileSetting) *ApplicationSetting {
	argument := NewArgumentSetting(nil)
	addition := NewAdditionSetting(nil)
	executor := NewExecutorSetting(nil)
	registry := NewRegistrySetting(nil)
	redirect := NewRedirectSetting(nil)
	for i := 0; i < len(profiles); i++ {
		profile := profiles[i]
		argument.Merge(profile.Argument)
		addition.Merge(profile.Addition)
		executor.Merge(profile.Executor)
		registry.Merge(profile.Registry)
		redirect.Merge(profile.Redirect)
	}
	executor.Merge(workspace.Setting.Executor)
	registry.Merge(workspace.Setting.Registry)
	redirect.Merge(workspace.Setting.Redirect)

	profile := &ApplicationSetting{
		Logger:         workspace.Logger,
		Workspace:      workspace,
		Argument:       argument,
		Addition:       addition,
		Executor:       executor,
		Registry:       registry,
		Redirect:       redirect,
		projectsByPath: map[string]*ProjectSetting{},
		projectsByName: map[string]*ProjectSetting{},
	}
	return profile
}

func (s *ApplicationSetting) GetAdditionProjectSettings(evaluator *Evaluator) ([]*ProjectSetting, error) {
	projectSettings, err := s.Addition.GetProjectSettings(evaluator)
	if err != nil {
		return nil, err
	}
	return projectSettings, nil
}

func (s *ApplicationSetting) GetExecutorItemSetting(name string) (*ExecutorItemSetting, error) {
	return s.Executor.GetItem(name, s.Workspace.Evaluator)
}

func (s *ApplicationSetting) GetRegistryLink(registry *ProjectLinkRegistry) (*ProjectLink, error) {
	evaluator := s.Workspace.Evaluator.SetRootData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ParsedRef.Type,
		"refName": registry.ParsedRef.Name,
	})
	return s.Registry.GetLink(registry.Name, evaluator)
}

func (s *ApplicationSetting) GetRedirectLink(resources []string) (*ProjectLink, string, error) {
	return s.Redirect.GetLink(resources, s.Workspace.Evaluator)
}

func (s *ApplicationSetting) GetProjectLinkTarget(link *ProjectLink) (target *ProjectLinkTarget, err error) {
	finalLink := link
	if link.Registry != nil {
		registryLink, err := s.GetRegistryLink(link.Registry)
		if err != nil {
			return nil, err
		}
		if registryLink == nil {
			return nil, ErrN("resolve project link error",
				Reason("registry not found"),
				KV("link", link),
			)
		}
		finalLink = registryLink
	}
	path := ""
	if finalLink.Dir != nil {
		path = finalLink.Dir.Path
	} else if finalLink.Git != nil {
		path = s.Workspace.GetGitProjectDir(finalLink.Git.ParsedUrl, finalLink.Git.ParsedRef)
	} else {
		Impossible()
	}
	resources := []string{
		finalLink.Normalized,
		"dir:" + path,
	}
	redirectLink, _, err := s.GetRedirectLink(resources)
	if err != nil {
		return nil, err
	}
	if redirectLink != nil {
		finalLink = redirectLink
		if finalLink.Dir != nil {
			path = finalLink.Dir.Path
		} else if finalLink.Git != nil {
			path = s.Workspace.GetGitProjectDir(finalLink.Git.ParsedUrl, finalLink.Git.ParsedRef)
		} else {
			Impossible()
		}
	}
	target = &ProjectLinkTarget{
		Link: link,
		Dir:  path,
		Git:  finalLink.Git,
	}
	return target, nil
}

func (s *ApplicationSetting) GetProjectEntityByRawLink(rawLink string) (*ProjectSetting, error) {
	link, err := ParseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	target, err := s.GetProjectLinkTarget(link)
	if err != nil {
		return nil, err
	}
	return s.GetProjectSettingByLinkTarget(target)
}

func (s *ApplicationSetting) GetProjectSettingByLinkTarget(target *ProjectLinkTarget) (*ProjectSetting, error) {
	if target.Git != nil {
		return s.getProjectEntityByGit(target.Dir, target.Git.Url, target.Git.ParsedUrl, target.Git.Ref, target.Git.ParsedRef)
	} else {
		return s.getProjectEntityByDir(target.Dir)
	}
}

func (s *ApplicationSetting) getProjectEntityByDir(path string) (*ProjectSetting, error) {
	if !IsDirExists(path) {
		return nil, ErrN("load project setting error",
			Reason("project dir not exists"),
			KV("path", path),
		)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, ErrW(err, "load project setting error",
			Reason("get abs-path error"),
			KV("path", path),
		)
	}
	path = absPath
	if setting, exist := s.projectsByPath[path]; exist {
		return setting, nil
	}

	s.Workspace.Logger.DebugDesc("load project setting", KV("path", path))
	var setting *ProjectSetting
	if setting, err = LoadProjectSetting(s.Workspace.Logger, path); err != nil {
		return nil, err
	}
	if existSetting, exist := s.projectsByName[setting.Name]; exist {
		if existSetting.Dir != setting.Dir {
			return nil, ErrN("get project setting error",
				Reason("project name duplicated"),
				KV("projectName", setting.Name),
				KV("projectPath1", setting.Dir),
				KV("projectPath2", existSetting.Dir),
			)
		}
	}
	s.projectsByPath[setting.Dir] = setting
	s.projectsByName[setting.Name] = setting
	return setting, nil
}

func (s *ApplicationSetting) getProjectEntityByGit(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *ProjectLinkGitRef) (entity *ProjectSetting, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, ErrW(err, "load project manifest error",
				Reason("parse url error"),
				KV("url", rawUrl),
				KV("ref", rawRef),
			)
		}
	}
	if parsedRef == nil {
		if parsedRef, err = ParseProjectLinkGitRef(rawRef); err != nil {
			return nil, ErrW(err, "load project manifest error",
				Reason("parse ref error"),
				KV("url", rawUrl),
				KV("ref", rawRef),
			)
		}
	}
	if path == "" {
		path = s.Workspace.GetGitProjectDir(parsedUrl, parsedRef)
	}
	if err = s.Workspace.DownloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, ErrW(err, "load project manifest error",
			Reason("download project error"),
			KV("url", rawUrl),
			KV("ref", rawRef),
		)
	}
	entity, err = s.getProjectEntityByDir(path)
	if err != nil {
		return nil, ErrW(err, "load project manifest error",
			Reason("load manifest error"),
			KV("url", rawUrl),
			KV("ref", rawRef),
		)
	}
	return entity, nil
}

func (s *ApplicationSetting) Inspect() *ApplicationSettingInspection {
	return NewApplicationSettingInspection(
		s.Argument.Inspect(),
		s.Addition.Inspect(),
		s.Executor.Inspect(),
		s.Registry.Inspect(),
		s.Redirect.Inspect(),
	)
}

// endregion
