package dsh_core

// region ProjectInspection

type ProjectInspection struct {
	Name   string                   `yaml:"name" toml:"name" json:"name"`
	Path   string                   `yaml:"path" toml:"path" json:"path"`
	Option *ProjectOptionInspection `yaml:"option" toml:"option" json:"option"`
	Script *ProjectScriptInspection `yaml:"script" toml:"script" json:"script"`
	Config *ProjectConfigInspection `yaml:"config" toml:"config" json:"config"`
}

func newProjectInspection(name string, path string, option *ProjectOptionInspection, script *ProjectScriptInspection, config *ProjectConfigInspection) *ProjectInspection {
	return &ProjectInspection{
		Name:   name,
		Path:   path,
		Option: option,
		Script: script,
		Config: config,
	}
}

// endregion
