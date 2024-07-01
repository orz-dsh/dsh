package inspection

// region ApplicationConfigInspection

type ApplicationConfigInspection struct {
	Value map[string]any `yaml:"value,omitempty" toml:"value,omitempty" json:"value,omitempty"`
	Trace map[string]any `yaml:"trace,omitempty" toml:"trace,omitempty" json:"trace,omitempty"`
}

func NewApplicationConfigInspection(value map[string]any, trace map[string]any) *ApplicationConfigInspection {
	return &ApplicationConfigInspection{
		Value: value,
		Trace: trace,
	}
}

// endregion
