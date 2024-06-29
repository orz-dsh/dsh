package internal

import (
	"fmt"
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// region ApplicationCore

type ApplicationCore struct {
	Logger                  *Logger
	Workspace               *WorkspaceCore
	Evaluator               *Evaluator
	Setting                 *ApplicationSetting
	Option                  *ApplicationOption
	MainProjectSetting      *ProjectSetting
	AdditionProjectSettings []*ProjectSetting
	MainProject             *Project
	AdditionProjects        []*Project
	DependencyProjects      []*Project
	Projects                []*Project
	projectsByName          map[string]*Project
	Configs                 map[string]any
	ConfigsTraces           map[string]any
}

func NewApplicationCore(workspace *WorkspaceCore, setting *ApplicationSetting, link string) (*ApplicationCore, error) {
	mainProjectSetting, err := setting.GetProjectEntityByRawLink(link)
	if err != nil {
		return nil, err
	}

	evaluator := workspace.Evaluator.SetData("main_project", map[string]any{
		"name": mainProjectSetting.Name,
		"path": mainProjectSetting.Dir,
	})

	arguments, err := setting.Argument.GetArguments(evaluator)
	if err != nil {
		return nil, err
	}

	additionProjectSettings, err := setting.GetAdditionProjectSettings(evaluator)
	if err != nil {
		return nil, err
	}

	option := NewApplicationOption(mainProjectSetting.Name, workspace.SystemInfo, evaluator, arguments)
	core := &ApplicationCore{
		Logger:                  workspace.Logger,
		Workspace:               workspace,
		Evaluator:               evaluator,
		Setting:                 setting,
		Option:                  option,
		MainProjectSetting:      mainProjectSetting,
		AdditionProjectSettings: additionProjectSettings,
		projectsByName:          map[string]*Project{},
	}
	return core, nil
}

func (a *ApplicationCore) loadProject(setting *ProjectSetting) (project *Project, err error) {
	if existProject, exist := a.projectsByName[setting.Name]; exist {
		return existProject, nil
	}
	if project, err = NewProject(a, setting, nil); err != nil {
		return nil, err
	}
	a.projectsByName[setting.Name] = project
	return project, nil
}

func (a *ApplicationCore) loadProjectByTarget(target *ProjectLinkTarget) (project *Project, err error) {
	setting, err := a.Setting.GetProjectSettingByLinkTarget(target)
	if err != nil {
		return nil, ErrW(err, "load project error",
			KV("reason", "load project setting error"),
			KV("target", target),
		)
	}
	project, err = a.loadProject(setting)
	if err != nil {
		return nil, ErrW(err, "load project error",
			KV("target", target),
		)
	}
	return project, nil
}

func (a *ApplicationCore) getProject(name string) *Project {
	return a.projectsByName[name]
}

func (a *ApplicationCore) loadImportProjects(project *Project, projectsDict map[string]bool) (projects []*Project, err error) {
	if err = project.loadImports(); err != nil {
		return nil, err
	}

	imp := project.import_
	for i := 0; i < len(imp.Items); i++ {
		p := imp.Items[i].project
		if !projectsDict[p.Dir] {
			projects = append(projects, p)
			projectsDict[p.Dir] = true
		}
	}

	for i := 0; i < len(projects); i++ {
		p1 := projects[i]
		if err = p1.loadImports(); err != nil {
			return nil, err
		}
		imp1 := p1.import_
		for j := 0; j < len(imp1.Items); j++ {
			p2 := imp1.Items[j].project
			if !projectsDict[p2.Dir] {
				projects = append(projects, p2)
				projectsDict[p2.Dir] = true
			}
		}
	}

	return projects, nil
}

func (a *ApplicationCore) loadProjects() (err error) {
	if a.MainProject != nil {
		return nil
	}

	// load main project
	var mainProject *Project
	if mainProject, err = a.loadProject(a.MainProjectSetting); err != nil {
		return err
	}
	var importProjects []*Project

	// load main project import projects
	projects := []*Project{mainProject}
	projectsDict := map[string]bool{mainProject.Dir: true}
	if _projects, err := a.loadImportProjects(mainProject, projectsDict); err != nil {
		return err
	} else {
		projects = append(projects, _projects...)
		importProjects = append(importProjects, _projects...)
	}

	// load extra projects
	var extraProjects []*Project
	for i := 0; i < len(a.AdditionProjectSettings); i++ {
		existProject := a.getProject(a.AdditionProjectSettings[i].Name)
		var option *ProjectOption
		if existProject != nil {
			option = existProject.option
		}
		extraProject, err := NewProject(a, a.AdditionProjectSettings[i], option)
		if err != nil {
			return err
		}
		extraProjects = append(extraProjects, extraProject)
		projects = append(projects, extraProject)

		// load extra project import projects
		if _projects, err := a.loadImportProjects(extraProject, projectsDict); err != nil {
			return err
		} else {
			projects = append(projects, _projects...)
			importProjects = append(importProjects, _projects...)
		}
	}

	a.MainProject = mainProject
	a.AdditionProjects = extraProjects
	a.DependencyProjects = importProjects
	a.Projects = projects
	return nil
}

func (a *ApplicationCore) MakeConfigs() (map[string]any, map[string]any, error) {
	if a.Configs != nil {
		return a.Configs, a.ConfigsTraces, nil
	}

	startTime := time.Now()
	a.Logger.Info("make configs start")

	if err := a.loadProjects(); err != nil {
		return nil, nil, ErrW(err, "make configs error",
			Reason("load projects error"),
			// TODO: error
		)
	}

	var contents []*ProjectResourceConfigItemContent
	for i := 0; i < len(a.Projects); i++ {
		iContents, err := a.Projects[i].loadConfigContents()
		if err != nil {
			return nil, nil, ErrW(err, "make configs error",
				Reason("load config contents error"),
				// TODO: error
				KV("project", a.Projects[i]),
			)
		}
		contents = append(contents, iContents...)
	}

	slices.SortStableFunc(contents, func(l, r *ProjectResourceConfigItemContent) int {
		n := l.Order - r.Order
		if n < 0 {
			return 1
		} else if n > 0 {
			return -1
		} else {
			return 0
		}
	})

	configs := map[string]any{}
	configsTraces := map[string]any{}
	for i := 0; i < len(contents); i++ {
		content := contents[i]
		if err := content.merge(configs, configsTraces); err != nil {
			return nil, nil, ErrW(err, "make configs error",
				Reason("merge configs error"),
				KV("file", content.file),
			)
		}
	}

	a.Configs = configs
	a.ConfigsTraces = configsTraces

	a.Logger.InfoDesc("make configs finish", KV("elapsed", time.Since(startTime)))
	return a.Configs, a.ConfigsTraces, nil
}

func (a *ApplicationCore) MakeArtifact(options MakeArtifactOptions) (artifact *ArtifactCore, err error) {
	configs, configsTraces, err := a.MakeConfigs()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	a.Logger.Info("make artifact start")
	outputDir := options.OutputDir
	if outputDir == "" {
		outputDir, err = a.Workspace.MakeOutputDir(a.MainProject.Name)
		if err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("make output path error"),
			)
		}
	} else {
		absPath, err := filepath.Abs(outputDir)
		if err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("get abs-path error"),
				KV("path", outputDir),
			)
		}
		outputDir = absPath
		if options.OutputDirClear {
			if err = ClearDir(outputDir); err != nil {
				return nil, ErrW(err, "make scripts error",
					Reason("clear output dir error"),
					KV("path", outputDir),
				)
			}
		}
	}

	evaluator := a.Evaluator.SetData("configs", configs).MergeFuncs(newProjectScriptTemplateFuncs())

	inspectionPath := ""
	if options.Inspection {
		inspectionPath = filepath.Join(outputDir, "@inspection")
		if err = os.MkdirAll(inspectionPath, os.ModePerm); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("make inspection dir error"),
				KV("path", inspectionPath),
			)
		}
		configsTracesInspectionPath := filepath.Join(inspectionPath, "app.configs-traces.yml")
		if err = WriteYamlFile(configsTracesInspectionPath, configsTraces); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("write configs traces inspection file error"),
				KV("path", configsTracesInspectionPath),
			)
		}
		dataInspectionPath := filepath.Join(inspectionPath, "app.data.yml")
		if err = WriteYamlFile(dataInspectionPath, evaluator.GetMap(false)); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("write data inspection file error"),
				KV("path", dataInspectionPath),
			)
		}
		optionInspectionPath := filepath.Join(inspectionPath, "app.option.yml")
		if err = WriteYamlFile(optionInspectionPath, a.Option.Inspect()); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("write option inspection file error"),
				KV("path", optionInspectionPath),
			)
		}
		profileInspectionPath := filepath.Join(inspectionPath, "app.setting.yml")
		if err = WriteYamlFile(profileInspectionPath, a.Setting.Inspect()); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("write profile inspection file error"),
				KV("path", profileInspectionPath),
			)
		}
	}

	if err := a.loadProjects(); err != nil {
		return nil, ErrW(err, "make scripts error",
			Reason("load projects error"),
			// TODO: error
		)
	}

	if inspectionPath != "" {
		mainProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.main.%s.yml", a.MainProject.Name))
		if err := WriteYamlFile(mainProjectInspectionPath, a.MainProject.inspect()); err != nil {
			return nil, ErrW(err, "make scripts error",
				Reason("write project inspection error"),
				KV("project", a.MainProject),
			)
		}
		for i := 0; i < len(a.AdditionProjects); i++ {
			extraProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.ext-%d.%s.yml", i+1, a.AdditionProjects[i].Name))
			if err := WriteYamlFile(extraProjectInspectionPath, a.AdditionProjects[i].inspect()); err != nil {
				return nil, ErrW(err, "make scripts error",
					Reason("write project inspection error"),
					KV("project", a.AdditionProjects[i]),
				)
			}
		}
		for i := 0; i < len(a.DependencyProjects); i++ {
			importProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.dep-%d.%s.yml", i+1, a.DependencyProjects[i].Name))
			if err := WriteYamlFile(importProjectInspectionPath, a.DependencyProjects[i].inspect()); err != nil {
				return nil, ErrW(err, "make scripts error",
					Reason("write project inspection error"),
					KV("project", a.DependencyProjects[i]),
				)
			}
		}
	}

	var targetNames []string
	var targetNamesDict = map[string]bool{}
	for i := 0; i < len(a.Projects); i++ {
		names, err := a.Projects[i].makeScripts(evaluator, outputDir, options.UseHardLink)
		if err != nil {
			return nil, err
		}
		for j := 0; j < len(names); j++ {
			if !targetNamesDict[names[j]] {
				targetNames = append(targetNames, names[j])
				targetNamesDict[names[j]] = true
			}
		}
	}

	a.Logger.InfoDesc("make artifact finish", KV("elapsed", time.Since(startTime)))
	return NewArtifactCore(a, outputDir, targetNames, targetNamesDict), nil
}

// endregion
