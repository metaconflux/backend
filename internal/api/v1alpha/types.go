package v1alpha

import "github.com/metaconflux/backend/internal/statics"

type MetadataSchema struct {
	Contract string             `json:"contract"`
	Static   statics.SpecSchema `json:"static"`
	Dynamic  []DynamicItem      `json:"dynamic"`
}

type DynamicItem struct {
	Target string      `json:"target"`
	Type   string      `json:"type"`
	Spec   interface{} `json:"spec"`
}

type MetadataResult struct {
	Url string `json:"url"`
}
