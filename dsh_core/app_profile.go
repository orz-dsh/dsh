package dsh_core

type appProfile struct {
	workspace                          *Workspace
	evaluator                          *Evaluator
	workspaceShellDefinitions          workspaceShellDefinitions
	workspaceImportRegistryDefinitions workspaceImportRegistryDefinitions
	workspaceImportRedirectDefinitions workspaceImportRedirectDefinitions
	projectOptionDefinitions           projectOptionDefinitions
	projectScriptSourceDefinitions     []*projectSourceDefinition
	projectScriptImportDefinitions     []*projectImportDefinition
	projectConfigSourceDefinitions     []*projectSourceDefinition
	projectConfigImportDefinitions     []*projectImportDefinition
}

func newAppProfile(workspace *Workspace, manifests []*AppProfileManifest) *appProfile {
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

	profile := &appProfile{
		workspace:                          workspace,
		evaluator:                          workspace.evaluator,
		workspaceShellDefinitions:          workspaceShellDefinitions,
		workspaceImportRegistryDefinitions: workspaceImportRegistryDefinitions,
		workspaceImportRedirectDefinitions: workspaceImportRedirectDefinitions,
		projectOptionDefinitions:           projectOptionDefinitions,
		projectScriptSourceDefinitions:     projectScriptSourceDefinitions,
		projectScriptImportDefinitions:     projectScriptImportDefinitions,
		projectConfigSourceDefinitions:     projectConfigSourceDefinitions,
		projectConfigImportDefinitions:     projectConfigImportDefinitions,
	}
	return profile
}

func (p *appProfile) getProjectOptions() (options map[string]string, err error) {
	options = make(map[string]string)
	if err = p.projectOptionDefinitions.fillOptions(options, p.evaluator); err != nil {
		return nil, err
	}
	return options, nil
}

func (p *appProfile) resolveProjectRawLink(rawLink string) (resolvedLink *projectResolvedLink, err error) {
	link, err := ParseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	resolvedLink, err = p.resolveProjectLink(link)
	if err != nil {
		return nil, err
	}
	return resolvedLink, nil
}

func (p *appProfile) resolveProjectLink(link *ProjectLink) (resolvedLink *projectResolvedLink, err error) {
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

func (p *appProfile) getWorkspaceShellDefinition(name string) (*workspaceShellDefinition, error) {
	definition := newWorkspaceShellDefinitionEmpty(name)
	if err := p.workspaceShellDefinitions.fillDefinition(definition, p.evaluator); err != nil {
		return nil, err
	}
	if err := definition.fillDefault(); err != nil {
		return nil, err
	}
	return definition, nil
}

func (p *appProfile) getWorkspaceImportRegistryLink(registry *ProjectLinkRegistry) (link *ProjectLink, err error) {
	evaluator := p.evaluator.SetRootData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ref.Type,
		"refName": registry.ref.Name,
	})
	link, err = p.workspaceImportRegistryDefinitions.getLink(registry.Name, evaluator)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (p *appProfile) getWorkspaceImportRedirectLink(resources []string) (link *ProjectLink, resource string, err error) {
	link, resource, err = p.workspaceImportRedirectDefinitions.getLink(resources, p.evaluator)
	if err != nil {
		return nil, "", err
	}
	return link, resource, nil
}
