package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region AdditionSettingModelBuilder

type AdditionSettingModelBuilder[P any] struct {
	commit func(*AdditionSettingModel) P
	items  []*AdditionItemSettingModel
}

func NewAdditionSettingModelBuilder[P any](commit func(*AdditionSettingModel) P) *AdditionSettingModelBuilder[P] {
	return &AdditionSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *AdditionSettingModelBuilder[P]) AddItemSetting(name, dir string) *AdditionItemSettingModelBuilder[*AdditionSettingModelBuilder[P]] {
	return NewAdditionItemSettingModelBuilder(b.addItemSettingModel, name, dir)
}

func (b *AdditionSettingModelBuilder[P]) CommitAdditionSetting() P {
	return b.commit(NewAdditionSettingModel(b.items))
}

func (b *AdditionSettingModelBuilder[P]) addItemSettingModel(item *AdditionItemSettingModel) *AdditionSettingModelBuilder[P] {
	b.items = append(b.items, item)
	return b
}

// endregion

// region AdditionItemSettingModelBuilder

type AdditionItemSettingModelBuilder[P any] struct {
	commit     func(*AdditionItemSettingModel) P
	name       string
	dir        string
	match      string
	dependency *ProjectDependencySettingModel
	resource   *ProjectResourceSettingModel
}

func NewAdditionItemSettingModelBuilder[P any](commit func(*AdditionItemSettingModel) P, name, dir string) *AdditionItemSettingModelBuilder[P] {
	return &AdditionItemSettingModelBuilder[P]{
		commit:     commit,
		name:       name,
		dir:        dir,
		dependency: NewProjectDependencySettingModel(nil),
		resource:   NewProjectResourceSettingModel(nil),
	}
}

func (b *AdditionItemSettingModelBuilder[P]) SetMatch(match string) *AdditionItemSettingModelBuilder[P] {
	b.match = match
	return b
}

func (b *AdditionItemSettingModelBuilder[P]) AddDependencyItem(link, match string) *AdditionItemSettingModelBuilder[P] {
	b.dependency.Items = append(b.dependency.Items, NewProjectDependencyItemSettingModel(link, match))
	return b
}

func (b *AdditionItemSettingModelBuilder[P]) AddResourceItem(dir string, includes, excludes []string, match string) *AdditionItemSettingModelBuilder[P] {
	b.resource.Items = append(b.resource.Items, NewProjectResourceItemSettingModel(dir, includes, excludes, match))
	return b
}

func (b *AdditionItemSettingModelBuilder[P]) CommitItemSetting() P {
	return b.commit(NewAdditionItemSettingModel(b.name, b.dir, b.match, b.dependency, b.resource))
}

// endregion
