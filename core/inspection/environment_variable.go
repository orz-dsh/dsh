package inspection

// region EnvironmentVariableInspection

type EnvironmentVariableInspection struct {
	Items []*EnvironmentVariableItemInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewEnvironmentVariableInspection(items []*EnvironmentVariableItemInspection) *EnvironmentVariableInspection {
	return &EnvironmentVariableInspection{
		Items: items,
	}
}

// endregion

// region EnvironmentVariableItemInspection

type EnvironmentVariableItemInspection struct {
	Key    string `yaml:"key" toml:"key" json:"key"`
	Name   string `yaml:"name" toml:"name" json:"name"`
	Value  string `yaml:"value" toml:"value" json:"value"`
	Source string `yaml:"source" toml:"source" json:"source"`
	Kind   string `yaml:"kind" toml:"kind" json:"kind"`
}

func NewEnvironmentVariableItemInspection(key, name, value, source, kind string) *EnvironmentVariableItemInspection {
	return &EnvironmentVariableItemInspection{
		Key:    key,
		Name:   name,
		Value:  value,
		Source: source,
		Kind:   kind,
	}
}

// endregion
