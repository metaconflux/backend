package container

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/metaconflux/backend/internal/containermanager"
	"github.com/metaconflux/backend/internal/gvk"
	"github.com/metaconflux/backend/internal/template"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
)

var _ transformers.ITransformer = &Transformer{}

var GVK = gvk.NewGroupVersionKind(
	"core",
	"v1alpha",
	"container",
)

type Transformer struct {
	cm     *containermanager.ContainerManager
	spec   SpecSchema
	params map[string]interface{}
	data   map[string]interface{}
}

// Validate implements transformers.ITransformer
func (t Transformer) Validate() error {
	return nil
}

type SpecSchema struct {
	Image string
}

func NewTransformer() (*Transformer, error) {
	cm, err := containermanager.NewManager("unix://run/user/1000/podman/podman.sock")
	if err != nil {
		return nil, err
	}
	return &Transformer{
		cm: cm,
	}, nil
}

// CreditsConsumed implements transformers.ITransformer
func (Transformer) CreditsConsumed() int {
	return 10
}

// Deadline implements transformers.ITransformer
func (Transformer) Deadline() time.Duration {
	return 10 * time.Second
}

// Execute implements transformers.ITransformer
func (t Transformer) Execute(ctx context.Context, base map[string]interface{}) (map[string]interface{}, error) {
	task := containermanager.Task{
		Timeout: t.Deadline() - 1*time.Second,
		Image:   t.spec.Image,
		Name:    uuid.NewString(),
		Input:   base,
	}

	result, err := t.cm.Task(task)
	if err != nil {
		return nil, err
	}

	var output map[string]interface{}
	err = utils.Remarshal(result, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Params implements transformers.ITransformer
func (Transformer) Params() map[string]interface{} {
	return make(map[string]interface{})
}

// Result implements transformers.ITransformer
func (Transformer) Result() interface{} {
	return nil
}

// Status implements transformers.ITransformer
func (Transformer) Status() []transformers.Status {
	return []transformers.Status{}
}

func (t Transformer) Prepare() error {
	cm, cancel := t.cm.WithTimeout(30 * time.Second)
	defer cancel()
	err := cm.PullIfNotPresent(nil, t.spec.Image)
	if err != nil {
		return err
	}
	return nil
}

// WithSpec implements transformers.ITransformer
func (t Transformer) WithSpec(ispec interface{}, params map[string]interface{}) (transformers.ITransformer, error) {
	var spec SpecSchema
	err := utils.Remarshal(ispec, &spec)
	if err != nil {
		return nil, err
	}

	transformer := Transformer{
		cm:     t.cm,
		params: params,
	}

	err = template.Template(&spec, &transformer.spec, params)
	if err != nil {
		return nil, err
	}

	return transformer, nil
}
