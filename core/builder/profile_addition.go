package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region ProfileAdditionSettingModelBuilder

type ProfileAdditionSettingModelBuilder[R any] struct {
	commit func(*ProfileAdditionSettingModel) R
	items  []*ProfileAdditionItemSettingModel
}

func NewProfileAdditionSettingModelBuilder[R any](commit func(*ProfileAdditionSettingModel) R) *ProfileAdditionSettingModelBuilder[R] {
	return &ProfileAdditionSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *ProfileAdditionSettingModelBuilder[R]) AddItemSetting(name, dir string) *ProfileAdditionItemSettingModelBuilder[*ProfileAdditionSettingModelBuilder[R]] {
	return NewProfileAdditionItemSettingModelBuilder(b.addItemSettingModel, name, dir)
}

func (b *ProfileAdditionSettingModelBuilder[R]) CommitAdditionSetting() R {
	return b.commit(NewProfileAdditionSettingModel(b.items))
}

func (b *ProfileAdditionSettingModelBuilder[R]) addItemSettingModel(item *ProfileAdditionItemSettingModel) *ProfileAdditionSettingModelBuilder[R] {
	b.items = append(b.items, item)
	return b
}

// endregion

// region ProfileAdditionItemSettingModelBuilder

type ProfileAdditionItemSettingModelBuilder[R any] struct {
	commit     func(*ProfileAdditionItemSettingModel) R
	name       string
	dir        string
	match      string
	dependency *ProjectDependencySettingModel
	resource   *ProjectResourceSettingModel
}

func NewProfileAdditionItemSettingModelBuilder[R any](commit func(*ProfileAdditionItemSettingModel) R, name, dir string) *ProfileAdditionItemSettingModelBuilder[R] {
	return &ProfileAdditionItemSettingModelBuilder[R]{
		commit:     commit,
		name:       name,
		dir:        dir,
		dependency: NewProjectDependencySettingModel(nil),
		resource:   NewProjectResourceSettingModel(nil),
	}
}

func (b *ProfileAdditionItemSettingModelBuilder[R]) SetMatch(match string) *ProfileAdditionItemSettingModelBuilder[R] {
	b.match = match
	return b
}

func (b *ProfileAdditionItemSettingModelBuilder[R]) AddDependencyItem(link, match string) *ProfileAdditionItemSettingModelBuilder[R] {
	b.dependency.Items = append(b.dependency.Items, NewProjectDependencyItemSettingModel(link, match))
	return b
}

func (b *ProfileAdditionItemSettingModelBuilder[R]) AddResourceItem(dir string, includes, excludes []string, match string) *ProfileAdditionItemSettingModelBuilder[R] {
	b.resource.Items = append(b.resource.Items, NewProjectResourceItemSettingModel(dir, includes, excludes, match))
	return b
}

func (b *ProfileAdditionItemSettingModelBuilder[R]) CommitItemSetting() R {
	return b.commit(NewProfileAdditionItemSettingModel(b.name, b.dir, b.match, b.dependency, b.resource))
}

// endregion
