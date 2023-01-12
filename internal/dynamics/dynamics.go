package dynamics

import "github.com/metaconflux/backend/internal/dynamics/contract"

type IDynamics interface {
	Get() (interface{}, error)
}

func NewDynamics(typ string, spec interface{}) (IDynamics, error) {
	var d IDynamics
	var err error

	switch typ {
	case contract.DYNAMICS_NAME:
		d, err = contract.NewDynamics(spec)
	}

	return d, err
}
