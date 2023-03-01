package print

import (
	"context"
	"time"

	"github.com/metaconflux/backend/internal/gvk"
	"github.com/metaconflux/backend/internal/template"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/sirupsen/logrus"
)

var GVK = gvk.NewGroupVersionKind(
	"core",
	"v1alpha",
	"print",
)

var deadline = 1 * time.Second
var _ transformers.ITransformer = &Transformer{}

type Transformer struct {
	spec   SpecSchema
	params map[string]interface{}
	data   map[string]interface{}
}

type SpecSchema struct {
	Something     string `json:"something" template:""`
	SomethingElse string `json:"somethingElse" template:""`
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
		params: params,
	}

	err = template.Template(&spec, &transformer.spec, params)
	if err != nil {
		return nil, err
	}

	return transformer, nil
}

func (t Transformer) Execute(ctx context.Context, base map[string]interface{}) (result map[string]interface{}, err error) {
	//time.Sleep(10 * time.Second)
	logrus.Infof("Some value: %s", t.spec.Something)
	logrus.Infof("Some other value: %s", t.spec.SomethingElse)
	//return nil, fmt.Errorf("Failed to print..kinda")

	return base, nil
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
	return nil
}

func (t Transformer) Prepare() error {
	return nil
}
