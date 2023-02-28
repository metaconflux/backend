package ipfs

import (
	"github.com/manifoldco/promptui"
	"github.com/metaconflux/backend/internal/transformers"
)

func NewSpecFromPrompt() (transformers.BaseTransformer, error) {
	var base transformers.BaseTransformer
	prompt := promptui.Prompt{
		Label:   "IPFS URL",
		Default: "ipfs://",
	}

	url, err := prompt.Run()
	if err != nil {
		return base, err
	}

	spec := SpecSchema{
		Url: url,
	}

	base = transformers.BaseTransformer{
		GroupVersionKind: GVK,
		Spec:             spec,
	}

	return base, nil

}
