package transformers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/metaconflux/backend/internal/gvk"
	"github.com/sirupsen/logrus"
)

var DEFAULT_DEADLINE = 5 * time.Second

type ITransformer interface {
	//Prepare() error
	//Transform(base interface{}) error
	WithSpec(spec interface{}, params map[string]interface{}) (ITransformer, error)
	Execute(base map[string]interface{}) (map[string]interface{}, error)
	Result() interface{}
	Status() []Status
	Params() map[string]interface{}
	CreditsConsumed() int
}

type NewTransformerFunc = func(spec interface{}, params map[string]interface{}) (ITransformer, error)

type TransformerInfo struct {
	New      NewTransformerFunc
	Credits  int
	Deadline time.Duration
}

type Status struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type ManifestInfo struct {
	GeneratedAt      time.Time `json:"generatedAt"`
	TransformerCount int       `json:"transformerCount"`
	Runtime          int64     `json:"runtime"`
	ManifestCID      string    `json:"manifestCID"`
}

type BaseTransformer struct {
	gvk.GroupVersionKind
	Spec   interface{} `json:"spec"`
	Status []Status    `json:"status"`
}

type Transformers struct {
	transformers map[gvk.GroupVersionKind]TransformerInfo
}

func NewTransformerManager() (*Transformers, error) {
	return &Transformers{
		transformers: make(map[gvk.GroupVersionKind]TransformerInfo),
	}, nil
}

func (t Transformers) Register(gvk gvk.GroupVersionKind, transformer NewTransformerFunc) error {
	if _, ok := t.transformers[gvk]; ok {
		return fmt.Errorf("Transformer %s already exists", gvk)
	}

	nt, err := transformer(nil, nil)
	if err != nil {
		return fmt.Errorf("Failed to initialize empty transformer: %s", err)
	}

	t.transformers[gvk] = TransformerInfo{
		New:      transformer,
		Credits:  nt.CreditsConsumed(),
		Deadline: DEFAULT_DEADLINE,
	}

	return nil
}

func (t Transformers) Get(gvk gvk.GroupVersionKind) (TransformerInfo, error) {
	transformer, ok := t.transformers[gvk]
	if !ok {
		return TransformerInfo{}, fmt.Errorf("Transformer %s unknown", gvk)
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

func (t Transformers) CalculateCredits(transformers []BaseTransformer) int {
	result := 0
	for _, tSpec := range transformers {
		ti, err := t.Get(tSpec.GroupVersionKind)
		if err != nil {
			return 0
		}

		result += ti.Credits
	}

	return result
}

func (t Transformers) Execute(transformers []BaseTransformer, params map[string]interface{}) (result map[string]interface{}, err error) {
	//results := make([]interface{}, len(transformers))
	start := time.Now()
	for _, tSpec := range transformers {
		func() {
			var ti TransformerInfo
			ti, err = t.Get(tSpec.GroupVersionKind)
			if err != nil {
				return
			}

			var transformer ITransformer
			transformer, err = ti.New(tSpec.Spec, params)
			if err != nil {
				return
			}
			defer func() {
				tSpec.Status = transformer.Status()
			}()

			result, err = transformer.Execute(result)
			if err != nil {
				logrus.Errorf("Failed to execute transformer %s: %s", tSpec.GroupVersionKind.String(), err)
				return
			}

			params["result"] = result

			err = t.UpdateParams(&params, transformer.Params())
			if err != nil {
				return
			}

		}()
		if err != nil {
			return
		}
	}

	end := time.Now()

	if err != nil {
		return
	}

	result["manifestInfo"] = ManifestInfo{
		GeneratedAt:      time.Now(),
		TransformerCount: len(transformers),
		Runtime:          end.Sub(start).Milliseconds(),
		ManifestCID:      params["manifestCID"].(string),
	}

	return
}
