package v1alpha

import (
	"github.com/metaconflux/backend/internal/hooks"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
)

type Manifest struct {
	Version      string                         `json:"version"`
	Owner        string                         `json:"owner"`
	Contract     string                         `json:"contract"`
	ChainID      int64                          `json:"chainId"`
	Transformers []transformers.BaseTransformer `json:"transformers"`
	Config       Config                         `json:"config"`
	Hooks        []hooks.Hook                   `json:"hooks"`
}

func (m Manifest) ValidVersion(version string) bool {
	return m.Version == version
}

type Config struct {
	Freeze       bool           `json:"freeze"`
	RefreshAfter utils.Duration `json:"refreshAfter"`
	Alias        string         `json:"alias"`
}

type DynamicItem struct {
	Target string      `json:"target"`
	Type   string      `json:"type"`
	Spec   interface{} `json:"spec"`
}

type MetadataResult struct {
	Url string `json:"url"`
}

type ManifestList struct {
	Address string `json:"address"`
	ChainId int64  `json:"chainId"`
	Alias   string `json:"alias"`
}
