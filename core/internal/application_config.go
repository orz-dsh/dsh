package internal

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"slices"
)

type ApplicationConfig struct {
	Value     map[string]any
	Trace     map[string]any
	Evaluator *Evaluator
}

func NewApplicationConfig(evaluator *Evaluator, projects []*Project) (*ApplicationConfig, error) {
	var contents []*ProjectResourceConfigItemContent
	for i := 0; i < len(projects); i++ {
		iContents, err := projects[i].loadConfigContents()
		if err != nil {
			return nil, ErrW(err, "make config error",
				Reason("load config contents error"),
				// TODO: error
				KV("project", projects[i]),
			)
		}
		contents = append(contents, iContents...)
	}

	slices.SortStableFunc(contents, func(l, r *ProjectResourceConfigItemContent) int {
		n := l.Order - r.Order
		if n < 0 {
			return 1
		} else if n > 0 {
			return -1
		} else {
			return 0
		}
	})

	value := map[string]any{}
	trace := map[string]any{}
	for i := 0; i < len(contents); i++ {
		content := contents[i]
		if err := content.merge(value, trace); err != nil {
			return nil, ErrW(err, "make config error",
				Reason("merge config error"),
				KV("file", content.file),
			)
		}
	}

	config := &ApplicationConfig{
		Value:     value,
		Trace:     trace,
		Evaluator: evaluator.SetData("config", value),
	}
	return config, nil
}

func (c *ApplicationConfig) Inspect() *ApplicationConfigInspection {
	return NewApplicationConfigInspection(c.Value, c.Trace)
}
