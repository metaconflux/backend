package local

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/metaconflux/backend/internal/gvk"
	"github.com/metaconflux/backend/internal/template"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
)

var GVK = gvk.NewGroupVersionKind(
	"core",
	"v1alpha",
	"local",
)

var deadline = 1 * time.Second
var _ transformers.ITransformer = &Transformer{}

type Transformer struct {
	spec        SpecSchema
	params      map[string]interface{}
	data        map[string]interface{}
	initialized bool
}

type SpecSchema struct {
	Path string `json:"path" template:""`
}

func NewTransformer() *Transformer {
	return &Transformer{}
}

func (t Transformer) WithSpec(ispec interface{}, params map[string]interface{}) (transformers.ITransformer, error) {
	var spec SpecSchema
	err := utils.Remarshal(ispec, &spec)
	if err != nil {
		return nil, err
	}

	transformer := Transformer{
		params:      params,
		initialized: true,
	}

	err = template.Template(&spec, &transformer.spec, params)
	if err != nil {
		return nil, err
	}

	return transformer, nil
}

func (t Transformer) Execute(ctx context.Context, base map[string]interface{}) (result map[string]interface{}, err error) {
	b, err := ioutil.ReadFile(t.spec.Path)
	if err != nil {
		return
	}

	err = json.Unmarshal(b, &result)
	if err != nil {
		return
	}

	return
}
func (t Transformer) Status() []transformers.Status {
	return nil
}

func (t Transformer) Params() map[string]interface{} {
	return t.params
}

func (t Transformer) Result() interface{} {
	return nil
}

func (t Transformer) CreditsConsumed() int {
	return 1
}

func (t Transformer) Deadline() time.Duration {
	return deadline
}

func (t Transformer) Validate() error {
	if !t.initialized {
		return fmt.Errorf("Not initialized")
	}

	if _, err := os.Stat(t.spec.Path); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Path does not exist")
	}
	return nil
}
