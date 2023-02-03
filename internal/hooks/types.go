package hooks

import "fmt"

type HookManager struct {
	hooks map[string]IHook
}

type IHook interface {
	WithSpec(spec interface{}, params map[string]interface{}) (IHook, error)
	Execute() error
}

type Hook struct {
	Type string      `json:"type"`
	Spec interface{} `json:"spec"`
}

func NewHooksManager() HookManager {
	return HookManager{
		hooks: make(map[string]IHook),
	}
}

func (h HookManager) Register(typ string, hook IHook) error {
	if _, ok := h.hooks[typ]; ok {
		return fmt.Errorf("Hook %s already registered", typ)
	}

	h.hooks[typ] = hook

	return nil
}

func (h HookManager) Get(typ string) (IHook, error) {
	hook, ok := h.hooks[typ]
	if !ok {
		return nil, fmt.Errorf("Hook %s unknown", typ)
	}

	return hook, nil
}
