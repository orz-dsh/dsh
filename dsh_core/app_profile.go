package dsh_core

type appProfile struct {
	workspace                       *Workspace
	evaluator                       *Evaluator
	workspaceShellEntities          workspaceShellEntitySet
	workspaceImportRegistryEntities workspaceImportRegistryEntitySet
	workspaceImportRedirectEntities workspaceImportRedirectEntitySet
	projectOptionDefinitions        projectOptionDefinitions
	projectScriptSourceDefinitions  []*projectSourceDefinition
	projectScriptImportDefinitions  []*projectImportDefinition
	projectConfigSourceDefinitions  []*projectSourceDefinition
	projectConfigImportDefinitions  []*projectImportDefinition
}

func newAppProfile(workspace *Workspace, manifests []*AppProfileManifest) *appProfile {
	workspaceShellEntities := workspaceShellEntitySet{}
	workspaceImportRegistryEntities := workspaceImportRegistryEntitySet{}
	workspaceImportRedirectEntities := workspaceImportRedirectEntitySet{}
	projectOptionDefinitions := projectOptionDefinitions{}
	projectScriptSourceDefinitions := []*projectSourceDefinition{}
	projectScriptImportDefinitions := []*projectImportDefinition{}
	projectConfigSourceDefinitions := []*projectSourceDefinition{}
	projectConfigImportDefinitions := []*projectImportDefinition{}
	for i := 0; i < len(manifests); i++ {
		manifest := manifests[i]
		workspaceShellEntities.merge(manifest.Workspace.Shell.entities)
		workspaceImportRegistryEntities.merge(manifest.Workspace.Import.registryDefinitions)
		workspaceImportRedirectEntities = append(workspaceImportRedirectEntities, manifest.Workspace.Import.redirectDefinitions...)
		projectOptionDefinitions = append(projectOptionDefinitions, manifest.Project.Option.definitions...)
		projectScriptSourceDefinitions = append(projectScriptSourceDefinitions, manifest.Project.Script.sourceDefinitions...)
		projectScriptImportDefinitions = append(projectScriptImportDefinitions, manifest.Project.Script.importDefinitions...)
		projectConfigSourceDefinitions = append(projectConfigSourceDefinitions, manifest.Project.Config.sourceDefinitions...)
		projectConfigImportDefinitions = append(projectConfigImportDefinitions, manifest.Project.Config.importDefinitions...)
	}
	workspaceShellEntities.merge(workspace.manifest.Shell.entities)
	workspaceShellEntities.mergeDefault()
	workspaceImportRegistryEntities.merge(workspace.manifest.Import.registryDefinitions)
	workspaceImportRegistryEntities.mergeDefault()
	workspaceImportRedirectEntities = append(workspaceImportRedirectEntities, workspace.manifest.Import.redirectDefinitions...)

	profile := &appProfile{
		workspace:                       workspace,
		evaluator:                       workspace.evaluator,
		workspaceShellEntities:          workspaceShellEntities,
		workspaceImportRegistryEntities: workspaceImportRegistryEntities,
		workspaceImportRedirectEntities: workspaceImportRedirectEntities,
		projectOptionDefinitions:        projectOptionDefinitions,
		projectScriptSourceDefinitions:  projectScriptSourceDefinitions,
		projectScriptImportDefinitions:  projectScriptImportDefinitions,
		projectConfigSourceDefinitions:  projectConfigSourceDefinitions,
		projectConfigImportDefinitions:  projectConfigImportDefinitions,
	}
	return profile
}

func (p *appProfile) getProjectOptionItems() (items map[string]string, err error) {
	return p.projectOptionDefinitions.getItems(p.evaluator)
}

func (p *appProfile) getWorkspaceShellEntity(name string) (*workspaceShellEntity, error) {
	return p.workspaceShellEntities.getEntity(name, p.evaluator)
}

func (p *appProfile) getWorkspaceImportRegistryLink(registry *ProjectLinkRegistry) (*ProjectLink, error) {
	evaluator := p.evaluator.SetRootData("registry", map[string]any{
		"name":    registry.Name,
		"path":    registry.Path,
		"ref":     registry.Ref,
		"refType": registry.ref.Type,
		"refName": registry.ref.Name,
	})
	return p.workspaceImportRegistryEntities.getLink(registry.Name, evaluator)
}

func (p *appProfile) getWorkspaceImportRedirectLink(resources []string) (*ProjectLink, string, error) {
	return p.workspaceImportRedirectEntities.getLink(resources, p.evaluator)
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
