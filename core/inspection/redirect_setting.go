package inspection

// region RedirectSettingInspection

type RedirectSettingInspection struct {
	Items []*RedirectItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewRedirectSettingInspection(items []*RedirectItemSettingInspection) *RedirectSettingInspection {
	return &RedirectSettingInspection{
		Items: items,
	}
}

// endregion

// region RedirectItemSettingInspection

type RedirectItemSettingInspection struct {
	Regex string `yaml:"regex" toml:"regex" json:"regex"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewRedirectItemSettingInspection(regex, link, match string) *RedirectItemSettingInspection {
	return &RedirectItemSettingInspection{
		Regex: regex,
		Link:  link,
		Match: match,
	}
}

// endregion
