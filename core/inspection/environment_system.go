package inspection

// region EnvironmentSystemInspection

type EnvironmentSystemInspection struct {
	Os         string            `yaml:"os" toml:"os" json:"os"`
	Arch       string            `yaml:"arch" toml:"arch" json:"arch"`
	Hostname   string            `yaml:"hostname" toml:"hostname" json:"hostname"`
	Username   string            `yaml:"username" toml:"username" json:"username"`
	HomeDir    string            `yaml:"homeDir" toml:"homeDir" json:"homeDir"`
	CurrentDir string            `yaml:"currentDir" toml:"currentDir" json:"currentDir"`
	Variables  map[string]string `yaml:"variables,omitempty" toml:"variables,omitempty" json:"variables,omitempty"`
}

func NewEnvironmentSystemInspection(os, arch, hostname, username, homeDir, currentDir string, variables map[string]string) *EnvironmentSystemInspection {
	return &EnvironmentSystemInspection{
		Os:         os,
		Arch:       arch,
		Hostname:   hostname,
		Username:   username,
		HomeDir:    homeDir,
		CurrentDir: currentDir,
		Variables:  variables,
	}
}

// endregion
