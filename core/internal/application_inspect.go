package internal

import (
	"fmt"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"path/filepath"
)

func (a *ApplicationCore) Inspect() (*ApplicationInspection, error) {
	if err := a.LoadConfig(); err != nil {
		// TODO: error
		return nil, err
	}
	if err := a.loadProjects(); err != nil {
		return nil, ErrW(err, "inspect application error",
			Reason("load projects error"),
			// TODO: error
		)
	}

	data := a.Evaluator.GetMap(false)

	var additionProjects []*ProjectInspection
	for i := 0; i < len(a.AdditionProjects); i++ {
		additionProjects = append(additionProjects, a.AdditionProjects[i].Inspect())
	}

	var dependencyProjects []*ProjectInspection
	for i := 0; i < len(a.DependencyProjects); i++ {
		dependencyProjects = append(dependencyProjects, a.DependencyProjects[i].Inspect())
	}

	inspection := NewApplicationInspection(
		data,
		a.Setting.Inspect(),
		a.Option.Inspect(),
		a.Config.Inspect(),
		a.MainProject.Inspect(),
		additionProjects,
		dependencyProjects,
	)
	return inspection, nil
}

func (a *ApplicationCore) SaveInspection(serializer Serializer, outputDir string) (err error) {
	inspection, err := a.Inspect()
	if err != nil {
		return ErrW(err, "make scripts error",
			Reason("inspect application error"),
		)
	}

	inspectionPath := filepath.Join(outputDir, "@inspection")
	if err = os.MkdirAll(inspectionPath, os.ModePerm); err != nil {
		return ErrW(err, "make scripts error",
			Reason("make inspection dir error"),
			KV("path", inspectionPath),
		)
	}

	dataInspectionPath := filepath.Join(inspectionPath, "app.data"+serializer.GetFileExt())
	if err = serializer.SerializeFile(dataInspectionPath, inspection.Data); err != nil {
		return ErrW(err, "make scripts error",
			Reason("write data inspection file error"),
			KV("path", dataInspectionPath),
		)
	}

	settingInspectionPath := filepath.Join(inspectionPath, "app.setting"+serializer.GetFileExt())
	if err = serializer.SerializeFile(settingInspectionPath, inspection.Setting); err != nil {
		return ErrW(err, "make scripts error",
			Reason("write setting inspection file error"),
			KV("path", settingInspectionPath),
		)
	}

	optionInspectionPath := filepath.Join(inspectionPath, "app.option"+serializer.GetFileExt())
	if err = serializer.SerializeFile(optionInspectionPath, inspection.Option); err != nil {
		return ErrW(err, "make scripts error",
			Reason("write option inspection file error"),
			KV("path", optionInspectionPath),
		)
	}

	configInspectionPath := filepath.Join(inspectionPath, "app.config"+serializer.GetFileExt())
	if err = serializer.SerializeFile(configInspectionPath, inspection.Config); err != nil {
		return ErrW(err, "make scripts error",
			Reason("write config inspection file error"),
			KV("path", configInspectionPath),
		)
	}

	mainProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.main.%s%s", inspection.MainProject.Name, serializer.GetFileExt()))
	if err = serializer.SerializeFile(mainProjectInspectionPath, inspection.MainProject); err != nil {
		return ErrW(err, "make scripts error",
			Reason("write project inspection error"),
			KV("project", inspection.MainProject.Name),
		)
	}

	for i := 0; i < len(inspection.AdditionProjects); i++ {
		extraProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.ext-%d.%s%s", i+1, inspection.AdditionProjects[i].Name, serializer.GetFileExt()))
		if err = serializer.SerializeFile(extraProjectInspectionPath, inspection.AdditionProjects[i]); err != nil {
			return ErrW(err, "make scripts error",
				Reason("write project inspection error"),
				KV("project", inspection.AdditionProjects[i].Name),
			)
		}
	}

	for i := 0; i < len(inspection.DependencyProjects); i++ {
		importProjectInspectionPath := filepath.Join(inspectionPath, fmt.Sprintf("project.dep-%d.%s%s", i+1, inspection.DependencyProjects[i].Name, serializer.GetFileExt()))
		if err = serializer.SerializeFile(importProjectInspectionPath, inspection.DependencyProjects[i]); err != nil {
			return ErrW(err, "make scripts error",
				Reason("write project inspection error"),
				KV("project", inspection.DependencyProjects[i].Name),
			)
		}
	}

	return nil
}
