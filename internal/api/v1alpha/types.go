package v1alpha

import (
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
)

type Manifest struct {
	Version      string                         `json:"version"`
	Owner        string                         `json:"owner"`
	Contract     string                         `json:"contract"`
	Transformers []transformers.BaseTransformer `json:"transformers"`
	Config       Config                         `json:"config"`
}

type Config struct {
	Freeze       bool           `json:"freeze"`
	RefreshAfter utils.Duration `json:"refreshAfter"`
}

type DynamicItem struct {
	Target string      `json:"target"`
	Type   string      `json:"type"`
	Spec   interface{} `json:"spec"`
}

type MetadataResult struct {
	Url string `json:"url"`
}

func (m Manifest) ValidVersion(version string) bool {
	return m.Version == version
}
