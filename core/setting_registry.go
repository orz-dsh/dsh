package core

// region default

var registrySettingDefault = newRegistrySetting([]*registryItemSetting{
	{
		Name: "orz-dsh",
		Link: "git:https://github.com/orz-dsh/{{.path}}.git#ref={{.ref}}",
	},
	{
		Name: "orz-ops",
		Link: "git:https://github.com/orz-ops/{{.path}}.git#ref={{.ref}}",
	},
})

// endregion

// region registrySetting

type registrySetting struct {
	Items       []*registryItemSetting
	itemsByName map[string][]*registryItemSetting
}

func newRegistrySetting(items []*registryItemSetting) *registrySetting {
	itemsByName := map[string][]*registryItemSetting{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		itemsByName[item.Name] = append(itemsByName[item.Name], item)
	}
	return &registrySetting{
		Items:       items,
		itemsByName: itemsByName,
	}
}

func (s *registrySetting) merge(setting *registrySetting) {
	for i := 0; i < len(setting.Items); i++ {
		item := setting.Items[i]
		s.Items = append(s.Items, item)
		s.itemsByName[item.Name] = append(s.itemsByName[item.Name], item)
	}
}

func (s *registrySetting) mergeDefault() {
	s.merge(registrySettingDefault)
}

func (s *registrySetting) getLink(name string, evaluator *Evaluator) (*projectLink, error) {
	if items, exist := s.itemsByName[name]; exist {
		for i := 0; i < len(items); i++ {
			model := items[i]
			matched, err := evaluator.EvalBoolExpr(model.match)
			if err != nil {
				return nil, errW(err, "get workspace import registry setting link error",
					reason("eval expr error"),
					kv("model", model),
				)
			}
			if matched {
				rawLink, err := evaluator.EvalStringTemplate(model.Link)
				if err != nil {
					return nil, errW(err, "get workspace import registry setting link error",
						reason("eval template error"),
						kv("model", model),
					)
				}
				link, err := parseProjectLink(rawLink)
				if err != nil {
					return nil, errW(err, "get workspace import registry setting link error",
						reason("parse link error"),
						kv("model", model),
						kv("rawLink", rawLink),
					)
				}
				return link, nil
			}
		}
	}
	return nil, nil
}

func (s *registrySetting) inspect() *WorkspaceRegistrySettingInspection {
	var items []*WorkspaceRegistryItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newWorkspaceRegistrySettingInspection(items)
}

// endregion

// region registryItemSetting

type registryItemSetting struct {
	Name  string
	Link  string
	Match string
	match *EvalExpr
}

func newRegistryItemSetting(name, link, match string, matchObj *EvalExpr) *registryItemSetting {
	return &registryItemSetting{
		Name:  name,
		Link:  link,
		Match: match,
		match: matchObj,
	}
}

func (s *registryItemSetting) inspect() *WorkspaceRegistryItemSettingInspection {
	return newWorkspaceRegistryItemSettingInspection(s.Name, s.Link, s.Match)
}

// endregion
