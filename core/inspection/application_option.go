package inspection

// region ApplicationOptionInspection

type ApplicationOptionInspection struct {
	Assign *ApplicationOptionAssignInspection `yaml:"assign" toml:"assign" json:"assign"`
	Common *ApplicationOptionCommonInspection `yaml:"common" toml:"common" json:"common"`
	Export *ApplicationOptionExportInspection `yaml:"export" toml:"export" json:"export"`
	Result *ApplicationOptionResultInspection `yaml:"result" toml:"result" json:"result"`
}

func NewApplicationOptionInspection(assign *ApplicationOptionAssignInspection, common *ApplicationOptionCommonInspection, export *ApplicationOptionExportInspection, result *ApplicationOptionResultInspection) *ApplicationOptionInspection {
	return &ApplicationOptionInspection{
		Assign: assign,
		Common: common,
		Export: export,
		Result: result,
	}
}

// endregion

// region ApplicationOptionAssignInspection

type ApplicationOptionAssignInspection struct {
	Common  map[string]string            `yaml:"common,omitempty" toml:"common,omitempty" json:"common,omitempty"`
	Export  map[string]string            `yaml:"export,omitempty" toml:"export,omitempty" json:"export,omitempty"`
	Project map[string]map[string]string `yaml:"project,omitempty" toml:"project,omitempty" json:"project,omitempty"`
}

func NewApplicationOptionAssignInspection(common map[string]string, export map[string]string, project map[string]map[string]string) *ApplicationOptionAssignInspection {
	return &ApplicationOptionAssignInspection{
		Common:  common,
		Export:  export,
		Project: project,
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
	Value  any                                          `yaml:"value" toml:"value" json:"value"`
	Type   string                                       `yaml:"type" toml:"type" json:"type"`
	Source string                                       `yaml:"source" toml:"source" json:"source"`
	Links  []*ApplicationOptionExportItemLinkInspection `yaml:"links,omitempty" toml:"links,omitempty" json:"links,omitempty"`
}

func NewApplicationOptionExportItemInspection(value any, typ, source string, links []*ApplicationOptionExportItemLinkInspection) *ApplicationOptionExportItemInspection {
	return &ApplicationOptionExportItemInspection{
		Value:  value,
		Type:   typ,
		Source: source,
		Links:  links,
	}
}

// endregion

// region ApplicationOptionExportItemLinkInspection

type ApplicationOptionExportItemLinkInspection struct {
	ProjectName string `yaml:"projectName" toml:"projectName" json:"projectName"`
	OptionName  string `yaml:"optionName" toml:"optionName" json:"optionName"`
}

func NewApplicationOptionExportItemLinkInspection(projectName, optionName string) *ApplicationOptionExportItemLinkInspection {
	return &ApplicationOptionExportItemLinkInspection{
		ProjectName: projectName,
		OptionName:  optionName,
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
