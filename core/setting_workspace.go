package core

import (
	"github.com/orz-dsh/dsh/utils"
	"path/filepath"
	"time"
)

// region default

var workspaceCleanSettingDefault = newWorkspaceCleanSetting(3, 24*time.Hour)

// endregion

// region workspaceSetting

type workspaceSetting struct {
	Clean    *workspaceCleanSetting
	Profile  *workspaceProfileSetting
	Executor *executorSetting
	Registry *registrySetting
	Redirect *redirectSetting
}

func newWorkspaceSetting(clean *workspaceCleanSetting, profile *workspaceProfileSetting, executor *executorSetting, registry *registrySetting, redirect *redirectSetting) *workspaceSetting {
	if clean == nil {
		clean = workspaceCleanSettingDefault
	}
	if profile == nil {
		profile = newWorkspaceProfileSetting(nil)
	}
	if executor == nil {
		executor = newExecutorSetting(nil)
	}
	if registry == nil {
		registry = newRegistrySetting(nil)
	}
	if redirect == nil {
		redirect = newRedirectSetting(nil)
	}
	return &workspaceSetting{
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func loadWorkspaceSetting(dir string) (setting *workspaceSetting, err error) {
	model := &workspaceSettingModel{}
	metadata, err := utils.DeserializeFromDir(dir, []string{"workspace"}, model, false)
	if err != nil {
		return nil, errW(err, "load workspace setting error",
			reason("deserialize error"),
			kv("dir", dir),
		)
	}
	file := ""
	if metadata != nil {
		file = metadata.File
	}
	if setting, err = model.convert(newModelHelper("workspace setting", file)); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region workspaceCleanSetting

type workspaceCleanSetting struct {
	OutputCount   int
	OutputExpires time.Duration
}

func newWorkspaceCleanSetting(outputCount int, outputExpires time.Duration) *workspaceCleanSetting {
	return &workspaceCleanSetting{
		OutputCount:   outputCount,
		OutputExpires: outputExpires,
	}
}

// endregion

// region workspaceProfileSetting

type workspaceProfileSetting struct {
	Items []*workspaceProfileItemSetting
}

func newWorkspaceProfileSetting(items []*workspaceProfileItemSetting) *workspaceProfileSetting {
	return &workspaceProfileSetting{
		Items: items,
	}
}

func (s *workspaceProfileSetting) getFiles(evaluator *Evaluator) ([]string, error) {
	var files []string
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if matched, err := evaluator.EvalBoolExpr(item.match); err != nil {
			return nil, errW(err, "get workspace profile setting files error",
				reason("eval expr error"),
				kv("item", item),
			)
		} else if matched {
			rawFile, err := evaluator.EvalStringTemplate(item.File)
			if err != nil {
				return nil, errW(err, "get workspace profile setting files error",
					reason("eval template error"),
					kv("item", item),
				)
			}
			file, err := filepath.Abs(rawFile)
			if err != nil {
				return nil, errW(err, "get workspace profile setting files error",
					reason("get abs-path error"),
					kv("item", item),
					kv("rawFile", rawFile),
				)
			}
			if utils.IsFileExists(file) {
				files = append(files, file)
			} else if !item.Optional {
				return nil, errN("get workspace profile setting files error",
					reason("file not found"),
					kv("item", item),
					kv("rawFile", rawFile),
					kv("file", file),
				)
			}
		}
	}
	return files, nil
}

// endregion

// region workspaceProfileItemSetting

type workspaceProfileItemSetting struct {
	File     string
	Optional bool
	Match    string
	match    *EvalExpr
}

func newWorkspaceProfileItemSetting(file string, optional bool, match string, matchObj *EvalExpr) *workspaceProfileItemSetting {
	return &workspaceProfileItemSetting{
		File:     file,
		Optional: optional,
		Match:    match,
		match:    matchObj,
	}
}

// endregion
