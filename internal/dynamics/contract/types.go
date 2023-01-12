package contract

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/metaconflux/backend/internal/utils"
)

const DYNAMICS_NAME = "contract"

type SpecSchema struct {
	Address  common.Address
	Function string
	Args     []interface{}
}

func NewDynamics(data interface{}) (SpecSchema, error) {
	var spec SpecSchema
	err := utils.Remarshal(data, &spec)
	if err != nil {
		return spec, err
	}

	return spec, nil
}

func (s SpecSchema) Get() (interface{}, error) {
	return "something", nil
}
