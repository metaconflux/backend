package print

import (
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

func init() {
	var _ transformers.ITransformer = &Transformer{}
}

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

func (t Transformer) Execute(base map[string]interface{}) (map[string]interface{}, error) {
	logrus.Infof("Some value: %s", t.spec.Something)
	logrus.Infof("Some other value: %s", t.spec.SomethingElse)

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
