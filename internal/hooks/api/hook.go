package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/metaconflux/backend/internal/hooks"
	"github.com/metaconflux/backend/internal/template"
	"github.com/metaconflux/backend/internal/utils"
)

const TYPE = "api"

type HookApi struct {
	hooks.IHook
	spec Spec
}

type Spec struct {
	Method string `json:"method"`
	Target string `json:"target" template:""`
	Status int    `json:"status"`
}

func NewHook() *HookApi {
	return &HookApi{}
}

func (h *HookApi) WithSpec(spec interface{}, params map[string]interface{}) (hooks.IHook, error) {
	hook := NewHook()
	specTmp := Spec{}
	err := utils.Remarshal(spec, &specTmp)
	if err != nil {
		return nil, err
	}

	err = template.Template(&specTmp, &hook.spec, params)
	if err != nil {
		return nil, err
	}

	return hook, nil
}

func (h *HookApi) Execute() error {
	log.Println("Trying hook ", h.spec)
	req, err := http.NewRequest(h.spec.Method, h.spec.Target, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != h.spec.Status {
		return fmt.Errorf("Unexpected response status %d", response.StatusCode)
	}

	return nil
}
