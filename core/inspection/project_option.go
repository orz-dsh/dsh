package inspection

// region ProjectOptionInspection

type ProjectOptionInspection struct {
	Items map[string]any `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProjectOptionInspection(items map[string]any) *ProjectOptionInspection {
	return &ProjectOptionInspection{
		Items: items,
	}
}

// endregion
