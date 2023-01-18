package contract

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/metaconflux/backend/internal/gvk"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/tidwall/sjson"
)

var GVK = gvk.NewGroupVersionKind(
	"core",
	"v1alpha",
	"contract",
)

func init() {
	var _ transformers.ITransformer = &Transformer{}
}

type Transformer struct {
	transformers.ITransformer
	spec    SpecSchema
	params  map[string]interface{}
	data    map[string]interface{}
	clients map[uint64]*w3.Client
}

type Arg struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Ret struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type SpecSchema struct {
	Address  common.Address `json:"address"`
	ChainID  uint64         `json:"chainId"`
	Function string         `json:"function"`
	Args     []Arg          `json:"args"`
	Returns  []Ret          `json:"returns"`
}

func (s SpecSchema) funcDef() string {
	args := make([]string, len(s.Args))

	for i, arg := range s.Args {
		args[i] = arg.Type
	}

	function := fmt.Sprintf("%s(%s)", s.Function, strings.Join(args, ","))

	return function
}

func (s SpecSchema) argValues() []any {
	result := make([]any, len(s.Args))

	for i, arg := range s.Args {
		switch arg.Type {
		case "uint256":
			val, err := strconv.Atoi(arg.Value.(string))
			if err != nil {
				return nil
			}
			result[i] = big.NewInt(int64(val))
		}
	}
	return result
}

func (s SpecSchema) retTypes() string {
	types := make([]string, len(s.Returns))

	for i, ret := range s.Returns {
		types[i] = ret.Type
	}

	return strings.Join(types, ",")
}

func (s SpecSchema) retTargets() []any {
	types := make([]any, len(s.Returns))

	for i, ret := range s.Returns {
		switch ret.Type {
		case "uint256":
			typ := big.NewInt(-1)
			types[i] = typ
		case "address":
			typ := common.Address{}
			types[i] = &typ
		}
		//TODO: Add remaining types
	}

	return types
}

func NewTransformer(clients map[uint64]*w3.Client) *Transformer {
	return &Transformer{
		clients: clients,
	}
}

func (t Transformer) WithSpec(ispec interface{}, params map[string]interface{}) (transformers.ITransformer, error) {
	var spec SpecSchema
	err := utils.Remarshal(ispec, &spec)
	if err != nil {
		return nil, err
	}

	for i, arg := range spec.Args {
		spec.Args[i].Value, err = utils.Template(arg.Value.(string), params)
		if err != nil {
			return nil, err
		}
	}

	log.Println(spec)

	return Transformer{
		spec:    spec,
		params:  params,
		clients: t.clients,
	}, nil
}

func (t Transformer) Execute(base map[string]interface{}) (map[string]interface{}, error) {
	w3func, err := w3.NewFunc(t.spec.funcDef(), t.spec.retTypes())
	if err != nil {
		return nil, err
	}

	data := t.spec.retTargets()
	log.Println(reflect.TypeOf(data[0]))

	err = t.clients[t.spec.ChainID].Call(
		eth.CallFunc(w3func, t.spec.Address, t.spec.argValues()...).Returns(data...),
	)
	if err != nil {
		return nil, err
	}

	result := ""

	baseBytes, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}

	for i, ret := range t.spec.Returns {
		result, err = sjson.Set(string(baseBytes), ret.Name, &(data[i]))
		if err != nil {
			return nil, err
		}

	}

	var resultMap map[string]interface{}

	err = json.Unmarshal([]byte(result), &resultMap)
	if err != nil {
		return nil, err
	}

	log.Println(resultMap)

	return resultMap, nil
}

func (t Transformer) Result() interface{} {
	return nil
}

func (t Transformer) Status() []transformers.Status {
	return nil
}

func (t Transformer) Params() map[string]interface{} {
	return nil
}
