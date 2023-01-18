package v1alpha

import (
	"github.com/metaconflux/backend/internal/transformers"
)

type MetadataSchema struct {
	Contract     string                         `json:"contract"`
	Transformers []transformers.BaseTransformer `json:"transformers"`
}

type DynamicItem struct {
	Target string      `json:"target"`
	Type   string      `json:"type"`
	Spec   interface{} `json:"spec"`
}

type MetadataResult struct {
	Url string `json:"url"`
}
