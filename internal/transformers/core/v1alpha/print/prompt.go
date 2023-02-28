package print

import (
	"github.com/manifoldco/promptui"
	"github.com/metaconflux/backend/internal/transformers"
)

func NewSpecFromPrompt() (transformers.BaseTransformer, error) {
	var base transformers.BaseTransformer
	prompt := promptui.Prompt{
		Label: "Say Something (will be templated!)",
	}

	something, err := prompt.Run()
	if err != nil {
		return base, err
	}

	prompt = promptui.Prompt{
		Label: "Say Something Else (also be templated)",
	}

	somethingElse, err := prompt.Run()
	if err != nil {
		return base, err
	}

	spec := SpecSchema{
		Something:     something,
		SomethingElse: somethingElse,
	}

	base = transformers.BaseTransformer{
		GroupVersionKind: GVK,
		Spec:             spec,
	}

	return base, nil

}
