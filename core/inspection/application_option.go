package inspection

// region ApplicationOptionInspection

type ApplicationOptionInspection struct {
	Common   *ApplicationOptionCommonInspection   `yaml:"common" toml:"common" json:"common"`
	Argument *ApplicationOptionArgumentInspection `yaml:"argument" toml:"argument" json:"argument"`
	Assign   *ApplicationOptionAssignInspection   `yaml:"assign" toml:"assign" json:"assign"`
	Result   *ApplicationOptionResultInspection   `yaml:"result" toml:"result" json:"result"`
}

func NewApplicationOptionInspection(common *ApplicationOptionCommonInspection, argument *ApplicationOptionArgumentInspection, assign *ApplicationOptionAssignInspection, result *ApplicationOptionResultInspection) *ApplicationOptionInspection {
	return &ApplicationOptionInspection{
		Common:   common,
		Argument: argument,
		Assign:   assign,
		Result:   result,
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

// region ApplicationOptionArgumentInspection

type ApplicationOptionArgumentInspection struct {
	Items map[string]map[string]string `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewApplicationOptionArgumentInspection(items map[string]map[string]string) *ApplicationOptionArgumentInspection {
	return &ApplicationOptionArgumentInspection{
		Items: items,
	}
}

// endregion

// region ApplicationOptionAssignInspection

type ApplicationOptionAssignInspection struct {
	Items map[string]*ApplicationOptionAssignItemInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewApplicationOptionAssignInspection(items map[string]*ApplicationOptionAssignItemInspection) *ApplicationOptionAssignInspection {
	return &ApplicationOptionAssignInspection{
		Items: items,
	}
}

// endregion

// region ApplicationOptionAssignItemInspection

type ApplicationOptionAssignItemInspection struct {
	Source      string `yaml:"source" toml:"source" json:"source"`
	FinalSource string `yaml:"finalSource" toml:"finalSource" json:"finalSource"`
	Mapping     string `yaml:"mapping,omitempty" toml:"mapping,omitempty" json:"mapping,omitempty"`
}

func NewApplicationOptionAssignItemInspection(source string, finalSource string, mapping string) *ApplicationOptionAssignItemInspection {
	return &ApplicationOptionAssignItemInspection{
		Source:      source,
		FinalSource: finalSource,
		Mapping:     mapping,
	}
}

// endregion

// region ApplicationOptionResultInspection

type ApplicationOptionResultInspection struct {
	Items map[string]*ApplicationOptionResultItemInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewApplicationOptionResultInspection(items map[string]*ApplicationOptionResultItemInspection) *ApplicationOptionResultInspection {
	return &ApplicationOptionResultInspection{
		Items: items,
	}
}

// endregion

// region ApplicationOptionResultItemInspection

type ApplicationOptionResultItemInspection struct {
	RawValue    string                                 `yaml:"rawValue" toml:"rawValue" json:"rawValue"`
	ParsedValue any                                    `yaml:"parsedValue" toml:"parsedValue" json:"parsedValue"`
	Source      string                                 `yaml:"source" toml:"source" json:"source"`
	Assign      *ApplicationOptionAssignItemInspection `yaml:"assign,omitempty" toml:"assign,omitempty" json:"assign,omitempty"`
}

func NewApplicationOptionResultItemInspection(rawValue string, parsedValue any, source string, assign *ApplicationOptionAssignItemInspection) *ApplicationOptionResultItemInspection {
	return &ApplicationOptionResultItemInspection{
		RawValue:    rawValue,
		ParsedValue: parsedValue,
		Source:      source,
		Assign:      assign,
	}
}

// endregion
