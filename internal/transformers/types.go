package transformers

import (
	"fmt"
	"reflect"

	"github.com/metaconflux/backend/internal/gvk"
)

type ITransformer interface {
	//Prepare() error
	//Transform(base interface{}) error
	WithSpec(spec interface{}, params map[string]interface{}) (ITransformer, error)
	Execute(base map[string]interface{}) (map[string]interface{}, error)
	Result() interface{}
	Status() []Status
	Params() map[string]interface{}
}

type NewTransformerFunc = func(spec interface{}, params map[string]interface{}) (ITransformer, error)

type Status struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type BaseTransformer struct {
	gvk.GroupVersionKind
	Spec   interface{} `json:"spec"`
	Status []Status    `json:"status"`
}

type Transformers struct {
	transformers map[gvk.GroupVersionKind]NewTransformerFunc
}

func NewTransformerManager() (*Transformers, error) {
	return &Transformers{
		transformers: make(map[gvk.GroupVersionKind]NewTransformerFunc),
	}, nil
}

func (t Transformers) Register(gvk gvk.GroupVersionKind, transformer NewTransformerFunc) error {
	if _, ok := t.transformers[gvk]; ok {
		return fmt.Errorf("Transformer %s already exists", gvk)
	}

	t.transformers[gvk] = transformer

	return nil
}

func (t Transformers) Get(gvk gvk.GroupVersionKind) (NewTransformerFunc, error) {
	transformer, ok := t.transformers[gvk]
	if !ok {
		return nil, fmt.Errorf("Transformer %s unknown", gvk)
	}

	return transformer, nil
}

func (t Transformers) UpdateParams(params *map[string]interface{}, toUpdate map[string]interface{}) error {
	locParams := *params
	for key, val := range locParams { //TODO: Nested stuff!!
		updateVal, ok := toUpdate[key]
		if !ok {
			continue
		}

		fieldType := reflect.TypeOf(val).Kind()
		fieldTypeUpdate := reflect.TypeOf(updateVal).Kind()

		if fieldType != fieldTypeUpdate {
			return fmt.Errorf("Cannot update non-matching types %s != %s", fieldType, fieldTypeUpdate)
		}

		switch fieldType {
		case reflect.Map:
			//t.UpdateParams(&val, updateVal)
		}

		(*params)[key] = updateVal
	}

	return nil
}

func (t Transformers) Execute(transformers []BaseTransformer, params map[string]interface{}) (result map[string]interface{}, err error) {
	//results := make([]interface{}, len(transformers))
	for _, tSpec := range transformers {
		func() {
			var newTransformer NewTransformerFunc
			newTransformer, err = t.Get(tSpec.GroupVersionKind)
			if err != nil {
				return
			}

			var transformer ITransformer
			transformer, err = newTransformer(tSpec.Spec, params)
			if err != nil {
				return
			}
			defer func() {
				tSpec.Status = transformer.Status()
			}()

			result, err = transformer.Execute(result)
			if err != nil {
				return
			}

			err = t.UpdateParams(&params, transformer.Params())
			if err != nil {
				return
			}

		}()
		if err != nil {
			return
		}
	}

	if err != nil {
		return
	}

	return
}
