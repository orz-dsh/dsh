package dsh_core

// region projectConfigSettingModel

type projectConfigSettingModel struct {
	Sources []*projectSourceSettingModel
	Imports []*projectImportSettingModel
}

func (m *projectConfigSettingModel) convert(ctx *modelConvertContext) (projectSourceSettingSet, projectImportSettingSet, error) {
	sourceSettings := projectSourceSettingSet{}
	for i := 0; i < len(m.Sources); i++ {
		if setting, err := m.Sources[i].convert(ctx.ChildItem("sources", i)); err != nil {
			return nil, nil, err
		} else {
			sourceSettings = append(sourceSettings, setting)
		}
	}

	importSettings := projectImportSettingSet{}
	for i := 0; i < len(m.Imports); i++ {
		if setting, err := m.Imports[i].convert(ctx.ChildItem("imports", i)); err != nil {
			return nil, nil, err
		} else {
			importSettings = append(importSettings, setting)
		}
	}

	return sourceSettings, importSettings, nil
}

// endregion
