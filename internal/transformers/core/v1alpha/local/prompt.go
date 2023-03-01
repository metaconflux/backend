package local

import (
	"github.com/manifoldco/promptui"
	"github.com/metaconflux/backend/internal/transformers"
)

func NewSpecFromPrompt() (transformers.BaseTransformer, error) {
	var base transformers.BaseTransformer
	prompt := promptui.Prompt{
		Label:   "Path to metadata files",
		Default: "./metadata/{{id}}.json",
	}

	path, err := prompt.Run()
	if err != nil {
		return base, err
	}

	spec := SpecSchema{
		Path: path,
	}

	base = transformers.BaseTransformer{
		GroupVersionKind: GVK,
		Spec:             spec,
	}

	return base, nil

}
