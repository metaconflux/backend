package transformers

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/metaconflux/backend/internal/gvk"
)

var DEFAULT_DEADLINE = 5 * time.Second

var ErrTransformerTimeout = fmt.Errorf("Transformer exceeded deadline")
var ErrTransformerUnknown = fmt.Errorf("Transformer unknown")

type ITransformer interface {
	//Prepare() error
	//Transform(base interface{}) error
	WithSpec(spec interface{}, params map[string]interface{}) (ITransformer, error)
	Execute(ctx context.Context, base map[string]interface{}) (map[string]interface{}, error)
	Validate() error
	Result() interface{}
	Status() []Status
	Params() map[string]interface{}
	CreditsConsumed() int
	Deadline() time.Duration
}

type NewTransformerFunc = func(spec interface{}, params map[string]interface{}) (ITransformer, error)
type NewSpecFromPrompt = func() (BaseTransformer, error)

type TransformerInfo struct {
	New      NewTransformerFunc
	Credits  int
	Deadline time.Duration
	Prompt   NewSpecFromPrompt
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
	Status []Status    `json:"status,omitempty"`
}

type Transformers struct {
	transformers map[gvk.GroupVersionKind]TransformerInfo
}

func NewTransformerManager() (*Transformers, error) {
	return &Transformers{
		transformers: make(map[gvk.GroupVersionKind]TransformerInfo),
	}, nil
}

func (t Transformers) Register(gvk gvk.GroupVersionKind, transformer NewTransformerFunc, prompt NewSpecFromPrompt) error {
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
		Deadline: nt.Deadline(),
		Prompt:   prompt,
	}

	return nil
}

func (t Transformers) Get(gvk gvk.GroupVersionKind) (TransformerInfo, error) {
	transformer, ok := t.transformers[gvk]
	if !ok {
		return TransformerInfo{}, ErrTransformerUnknown
	}

	return transformer, nil
}

func (t Transformers) GetRegistered() []gvk.GroupVersionKind {
	keys := make([]gvk.GroupVersionKind, 0, len(t.transformers))
	for key := range t.transformers {
		keys = append(keys, key)
	}

	return keys
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

			ctx, cancel := context.WithTimeout(context.Background(), ti.Deadline)
			defer cancel()

			resultCh := make(chan map[string]interface{}, 1)
			errCh := make(chan error, 1)

			go func() {
				result, err := transformer.Execute(ctx, result)
				if err != nil {
					errCh <- err
					return
				}

				resultCh <- result
			}()

			select {
			case <-ctx.Done():
				err = fmt.Errorf("%s: %s", tSpec.GroupVersionKind.String(), ErrTransformerTimeout)
				return
			case r := <-resultCh:
				result = r
			case e := <-errCh:
				err = fmt.Errorf("%s: %s", tSpec.GroupVersionKind.String(), e)
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

	manifestCID := ""
	if _, ok := params["manifestCID"]; ok {
		manifestCID = params["manifestCID"].(string)
	}

	result["manifestInfo"] = ManifestInfo{
		GeneratedAt:      time.Now(),
		TransformerCount: len(transformers),
		Runtime:          end.Sub(start).Milliseconds(),
		ManifestCID:      manifestCID,
	}

	return
}

func (t Transformers) Validate(transformers []BaseTransformer) error {
	var failedTransformers []string
	var failedValidations []string
	for _, ts := range transformers {
		tf, err := t.Get(ts.GroupVersionKind)
		if errors.Is(err, ErrTransformerUnknown) {
			failedTransformers = append(failedTransformers, ts.GroupVersionKind.String())
			continue
		}

		transformer, err := tf.New(ts.Spec, map[string]interface{}{})
		if err != nil {
			failedValidations = append(failedValidations, fmt.Sprintf("Failed to instantiate %s: %s", ts.GroupVersionKind.String(), err))
		}

		err = transformer.Validate()
		if err != nil {
			failedValidations = append(failedValidations, fmt.Sprintf("Failed to validate %s: %s", ts.GroupVersionKind.String(), err))
		}
	}

	var result string
	if len(failedTransformers) > 0 {
		result += fmt.Sprintf("Transformers %s not registered\n", strings.Join(failedTransformers, ", "))
	}

	if len(failedValidations) > 0 {
		result += fmt.Sprintf("Validation failed: %s", strings.Join(failedValidations, ", "))
	}

	if len(result) > 0 {
		return fmt.Errorf(result)
	}

	return nil
}
