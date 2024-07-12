package inspection

// region ApplicationOptionInspection

type ApplicationOptionInspection struct {
	Common *ApplicationOptionCommonInspection `yaml:"common" toml:"common" json:"common"`
	Export *ApplicationOptionExportInspection `yaml:"export" toml:"export" json:"export"`
	Assign *ApplicationOptionAssignInspection `yaml:"assign" toml:"assign" json:"assign"`
	Result *ApplicationOptionResultInspection `yaml:"result" toml:"result" json:"result"`
}

func NewApplicationOptionInspection(common *ApplicationOptionCommonInspection, export *ApplicationOptionExportInspection, assign *ApplicationOptionAssignInspection, result *ApplicationOptionResultInspection) *ApplicationOptionInspection {
	return &ApplicationOptionInspection{
		Common: common,
		Export: export,
		Assign: assign,
		Result: result,
	}
}

// endregion

// region ApplicationOptionCommonInspection

type ApplicationOptionCommonInspection struct {
	Os       string `yaml:"os" toml:"os" json:"os"`
	Arch     string `yaml:"arch" toml:"arch" json:"arch"`
	Executor string `yaml:"executor" toml:"executor" json:"executor"`
	Hostname string `yaml:"hostname" toml:"hostname" json:"hostname"`
	Username string `yaml:"username" toml:"username" json:"username"`
}

func NewApplicationOptionCommonInspection(os, arch, executor, hostname, username string) *ApplicationOptionCommonInspection {
	return &ApplicationOptionCommonInspection{
		Os:       os,
		Arch:     arch,
		Executor: executor,
		Hostname: hostname,
		Username: username,
	}
}

// endregion

// region ApplicationOptionExportInspection

type ApplicationOptionExportInspection struct {
	Items map[string]*ApplicationOptionExportItemInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewApplicationOptionExportInspection(items map[string]*ApplicationOptionExportItemInspection) *ApplicationOptionExportInspection {
	return &ApplicationOptionExportInspection{
		Items: items,
	}
}

// endregion

// region ApplicationOptionExportItemInspection

type ApplicationOptionExportItemInspection struct {
	Value  any    `yaml:"value" toml:"value" json:"value"`
	Type   string `yaml:"type" toml:"type" json:"type"`
	Source string `yaml:"source" toml:"source" json:"source"`
}

func NewApplicationOptionExportItemInspection(value any, typ, source string) *ApplicationOptionExportItemInspection {
	return &ApplicationOptionExportItemInspection{
		Value:  value,
		Type:   typ,
		Source: source,
	}
}

// endregion

// region ApplicationOptionAssignInspection

type ApplicationOptionAssignInspection struct {
	Items map[string]map[string]string `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewApplicationOptionAssignInspection(items map[string]map[string]string) *ApplicationOptionAssignInspection {
	return &ApplicationOptionAssignInspection{
		Items: items,
	}
}

// endregion

// region ApplicationOptionResultInspection

type ApplicationOptionResultInspection struct {
	Items map[string]map[string]*ApplicationOptionResultItemInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewApplicationOptionResultInspection(items map[string]map[string]*ApplicationOptionResultItemInspection) *ApplicationOptionResultInspection {
	return &ApplicationOptionResultInspection{
		Items: items,
	}
}

// endregion

// region ApplicationOptionResultItemInspection

type ApplicationOptionResultItemInspection struct {
	Value  any    `yaml:"value" toml:"value" json:"value"`
	Source string `yaml:"source" toml:"source" json:"source"`
}

func NewApplicationOptionResultItemInspection(value any, source string) *ApplicationOptionResultItemInspection {
	return &ApplicationOptionResultItemInspection{
		Value:  value,
		Source: source,
	}
}

// endregion
