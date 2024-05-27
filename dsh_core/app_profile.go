package dsh_core

import (
	"dsh/dsh_utils"
)

type AppProfile struct {
	workspace                          *Workspace
	evalData                           *appEvalData
	workspaceShellDefinitions          workspaceShellDefinitions
	workspaceImportRegistryDefinitions workspaceImportRegistryDefinitions
	workspaceImportRedirectDefinitions workspaceImportRedirectDefinitions
	projectOptionDefinitions           projectOptionDefinitions
	projectScriptSourceDefinitions     []*projectSourceDefinition
	projectScriptImportDefinitions     []*projectImportDefinition
	projectConfigSourceDefinitions     []*projectSourceDefinition
	projectConfigImportDefinitions     []*projectImportDefinition
}

func makeAppProfile(workspace *Workspace, evalData *appEvalData, manifests []*AppProfileManifest) (*AppProfile, error) {
	workspaceShellDefinitions := workspaceShellDefinitions{}
	workspaceImportRegistryDefinitions := workspaceImportRegistryDefinitions{}
	workspaceImportRedirectDefinitions := workspaceImportRedirectDefinitions{}
	projectOptionDefinitions := projectOptionDefinitions{}
	projectScriptSourceDefinitions := []*projectSourceDefinition{}
	projectScriptImportDefinitions := []*projectImportDefinition{}
	projectConfigSourceDefinitions := []*projectSourceDefinition{}
	projectConfigImportDefinitions := []*projectImportDefinition{}
	for i := 0; i < len(manifests); i++ {
		manifest := manifests[i]
		for name, definitions := range manifest.Workspace.Shell.definitions {
			workspaceShellDefinitions[name] = append(workspaceShellDefinitions[name], definitions...)
		}
		for name, definitions := range manifest.Workspace.Import.registryDefinitions {
			workspaceImportRegistryDefinitions[name] = append(workspaceImportRegistryDefinitions[name], definitions...)
		}
		workspaceImportRedirectDefinitions = append(workspaceImportRedirectDefinitions, manifest.Workspace.Import.redirectDefinitions...)
		projectOptionDefinitions = append(projectOptionDefinitions, manifest.Project.Option.definitions...)
		projectScriptSourceDefinitions = append(projectScriptSourceDefinitions, manifest.Project.Script.sourceDefinitions...)
		projectScriptImportDefinitions = append(projectScriptImportDefinitions, manifest.Project.Script.importDefinitions...)
		projectConfigSourceDefinitions = append(projectConfigSourceDefinitions, manifest.Project.Config.sourceDefinitions...)
		projectConfigImportDefinitions = append(projectConfigImportDefinitions, manifest.Project.Config.importDefinitions...)
	}
	for name, definitions := range workspace.manifest.Shell.definitions {
		workspaceShellDefinitions[name] = append(workspaceShellDefinitions[name], definitions...)
	}
	for name, definitions := range workspace.manifest.Import.registryDefinitions {
		workspaceImportRegistryDefinitions[name] = append(workspaceImportRegistryDefinitions[name], definitions...)
	}
	for name, definitions := range workspaceImportRegistryDefinitionsDefault {
		workspaceImportRegistryDefinitions[name] = append(workspaceImportRegistryDefinitions[name], definitions...)
	}
	workspaceImportRedirectDefinitions = append(workspaceImportRedirectDefinitions, workspace.manifest.Import.redirectDefinitions...)

	profile := &AppProfile{
		workspace:                          workspace,
		evalData:                           evalData,
		workspaceShellDefinitions:          workspaceShellDefinitions,
		workspaceImportRegistryDefinitions: workspaceImportRegistryDefinitions,
		workspaceImportRedirectDefinitions: workspaceImportRedirectDefinitions,
		projectOptionDefinitions:           projectOptionDefinitions,
		projectScriptSourceDefinitions:     projectScriptSourceDefinitions,
		projectScriptImportDefinitions:     projectScriptImportDefinitions,
		projectConfigSourceDefinitions:     projectConfigSourceDefinitions,
		projectConfigImportDefinitions:     projectConfigImportDefinitions,
	}
	return profile, nil
}

func (p *AppProfile) getOptionValues() (options map[string]string, err error) {
	evalData := p.evalData
	matcher := dsh_utils.NewEvalMatcher(evalData)
	options = make(map[string]string)
	if err = p.projectOptionDefinitions.fillOptions(options, matcher); err != nil {
		return nil, err
	}
	return options, nil
}

func (p *AppProfile) resolveProjectLink(link *ProjectLink) (resolvedLink *projectResolvedLink, err error) {
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
	resolvedLink = &projectResolvedLink{
		Link: link,
		Path: path,
		Git:  finalLink.Git,
	}
	return resolvedLink, nil
}

func (p *AppProfile) getWorkspaceShellDefinition(name string) (*workspaceShellDefinition, error) {
	evalData := p.evalData
	matcher := dsh_utils.NewEvalMatcher(evalData)
	definition := newWorkspaceShellDefinitionEmpty(name)
	if err := p.workspaceShellDefinitions.fillDefinition(definition, matcher); err != nil {
		return nil, err
	}
	if err := definition.fillDefault(); err != nil {
		return nil, err
	}
	return definition, nil
}

func (p *AppProfile) getWorkspaceImportRegistryLink(registry *ProjectLinkRegistry) (link *ProjectLink, err error) {
	evalData := p.evalData.MainData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ref.Type,
		"refName": registry.ref.Name,
	})
	matcher := dsh_utils.NewEvalMatcher(evalData)
	replacer := dsh_utils.NewEvalReplacer(evalData, nil)
	link, err = p.workspaceImportRegistryDefinitions.getLink(registry.Name, matcher, replacer)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (p *AppProfile) getWorkspaceImportRedirectLink(resources []string) (link *ProjectLink, resource string, err error) {
	evalData := p.evalData
	matcher := dsh_utils.NewEvalMatcher(evalData)
	replacer := dsh_utils.NewEvalReplacer(evalData, nil)
	link, resource, err = p.workspaceImportRedirectDefinitions.getLink(resources, matcher, replacer)
	if err != nil {
		return nil, "", err
	}
	return link, resource, nil
}
