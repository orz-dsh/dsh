package internal

import (
	. "github.com/orz-dsh/dsh/core/builder"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
)

// region EnvironmentCore

type EnvironmentCore struct {
	Logger    *Logger
	System    *System
	Variable  *EnvironmentVariable
	Setting   *EnvironmentSetting
	Evaluator *Evaluator
}

func NewEnvironmentCore(logger *Logger, assigns map[string]string) (*EnvironmentCore, error) {
	system, err := GetSystem()
	if err != nil {
		return nil, ErrW(err, "new environment error",
			Reason("get system error"),
		)
	}
	core := &EnvironmentCore{
		Logger:   logger,
		System:   system,
		Variable: NewEnvironmentVariable(system, assigns),
	}

	setting, err := core.BuildSetting()
	if err != nil {
		return nil, ErrW(err, "new environment error",
			Reason("build setting error"),
		)
	}
	core.Setting = setting
	core.Evaluator = NewEvaluator().
		SetData("local", map[string]any{
			"os":                   system.Os,
			"arch":                 system.Arch,
			"hostname":             system.Hostname,
			"username":             system.Username,
			"home_dir":             system.HomeDir,
			"current_dir":          system.CurrentDir,
			"runtime_version":      GetRuntimeVersion(),
			"runtime_version_code": GetRuntimeVersionCode(),
		}).
		SetData("global", MapAnyByStr(setting.Argument.GetMap()))

	return core, nil
}

func (e *EnvironmentCore) BuildSetting() (*EnvironmentSetting, error) {
	builder := NewEnvironmentSettingModelBuilder(func(model *EnvironmentSettingModel) *EnvironmentSettingModel {
		return model
	})
	argumentBuilder := builder.SetArgumentSetting()
	workspaceBuilder := builder.SetWorkspaceSetting()
	workspaceProfileItems := EnvironmentVariableParsedItemSlice[*WorkspaceProfileItemSettingModel]{}
	workspaceExecutorItems := EnvironmentVariableParsedItemSlice[*ExecutorItemSettingModel]{}
	workspaceRegistryItems := EnvironmentVariableParsedItemSlice[*RegistryItemSettingModel]{}
	workspaceRedirectItems := EnvironmentVariableParsedItemSlice[*RedirectItemSettingModel]{}

	for i := 0; i < len(e.Variable.Items); i++ {
		item := e.Variable.Items[i]
		switch item.Kind {
		case EnvironmentVariableKindArgumentItem:
			argumentBuilder.AddItem(item.Name, item.Value)
		case EnvironmentVariableKindWorkspaceDir:
			workspaceBuilder.SetDir(item.Value)
		case EnvironmentVariableKindWorkspaceClean:
			parsed, err := NewEnvironmentVariableParsedItem(item, &WorkspaceCleanSettingModel{})
			if err != nil {
				return nil, err
			}
			workspaceBuilder.SetCleanSetting().SetOutputModel(parsed.Value.Output).CommitCleanSetting()
		case EnvironmentVariableKindWorkspaceProfile:
			parsed, err := NewEnvironmentVariableParsedItem(item, &WorkspaceProfileItemSettingModel{})
			if err != nil {
				return nil, err
			}
			workspaceProfileItems = append(workspaceProfileItems, parsed)
		case EnvironmentVariableKindWorkspaceExecutor:
			parsed, err := NewEnvironmentVariableParsedItem(item, &ExecutorItemSettingModel{})
			if err != nil {
				return nil, err
			}
			workspaceExecutorItems = append(workspaceExecutorItems, parsed)
		case EnvironmentVariableKindWorkspaceRegistry:
			parsed, err := NewEnvironmentVariableParsedItem(item, &RegistryItemSettingModel{})
			if err != nil {
				return nil, err
			}
			workspaceRegistryItems = append(workspaceRegistryItems, parsed)
		case EnvironmentVariableKindWorkspaceRedirect:
			parsed, err := NewEnvironmentVariableParsedItem(item, &RedirectItemSettingModel{})
			if err != nil {
				return nil, err
			}
			workspaceRedirectItems = append(workspaceRedirectItems, parsed)
		default:
			e.Logger.WarnDesc("environment variable unknown", KV("item", item))
		}
	}

	argumentBuilder.CommitArgumentSetting()
	workspaceBuilder.
		SetProfileSetting().SetItems(workspaceProfileItems.Sort().GetValues()).CommitProfileSetting().
		SetExecutorSetting().SetItems(workspaceExecutorItems.Sort().GetValues()).CommitExecutorSetting().
		SetRegistrySetting().SetItems(workspaceRegistryItems.Sort().GetValues()).CommitRegistrySetting().
		SetRedirectSetting().SetItems(workspaceRedirectItems.Sort().GetValues()).CommitRedirectSetting().
		CommitWorkspaceSetting()

	model := builder.CommitEnvironmentSetting()
	setting, err := model.Convert(NewModelHelper(nil, "environment setting", "environment"))
	if err != nil {
		return nil, err
	}
	return setting, nil
}

func (e *EnvironmentCore) GetWorkspaceDir() string {
	if e.Setting.Workspace.Dir != "" {
		return e.Setting.Workspace.Dir
	}
	return filepath.Join(e.System.HomeDir, "dsh")
}

func (e *EnvironmentCore) Inspect() *EnvironmentInspection {
	return NewEnvironmentInspection(
		NewEnvironmentSystemInspection(
			e.System.Os,
			e.System.Arch,
			e.System.Hostname,
			e.System.Username,
			e.System.HomeDir,
			e.System.CurrentDir,
			e.System.Variables,
		),
		e.Variable.Inspect(),
		e.Setting.Inspect(),
	)
}

// endregion
