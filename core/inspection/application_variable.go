package inspection

// region ApplicationVariableInspection

type ApplicationVariableInspection struct {
	Local  map[string]any `yaml:"local,omitempty" toml:"local,omitempty" json:"local,omitempty"`
	Global map[string]any `yaml:"global,omitempty" toml:"global,omitempty" json:"global,omitempty"`
}

func NewApplicationVariableInspection(local map[string]any, global map[string]any) *ApplicationVariableInspection {
	return &ApplicationVariableInspection{
		Local:  local,
		Global: global,
	}
}

// endregion
