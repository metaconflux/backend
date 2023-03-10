package contract

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
	"github.com/metaconflux/backend/internal/gvk"
	"github.com/metaconflux/backend/internal/template"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/tidwall/sjson"
)

var GVK = gvk.NewGroupVersionKind(
	"core",
	"v1alpha",
	"contract",
)

var deadline = 3 * time.Second
var _ transformers.ITransformer = &Transformer{}

type Transformer struct {
	spec    SpecSchema
	params  map[string]interface{}
	data    map[string]interface{}
	clients map[uint64]*w3.Client
}

type Arg struct {
	Type  string `json:"type" template:""`
	Value string `json:"value" template:""`
}

type Ret struct {
	Name string `json:"name" template:""`
	Type string `json:"type" template:""`
}

type SpecSchema struct {
	Address  common.Address `json:"address" template:""`
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

func (s SpecSchema) argValues() ([]any, error) {
	result := make([]any, len(s.Args))

	for i, arg := range s.Args {
		switch arg.Type {
		case "uint256", "int256", "uint", "int":
			val, err := strconv.Atoi(arg.Value)
			if err != nil {
				return nil, err
			}
			result[i] = big.NewInt(int64(val))
		case "address":
			val := common.HexToAddress(arg.Value)
			result[i] = val
		case "string":
			result[i] = arg.Value
		case "bool":
			val, err := strconv.ParseBool(arg.Value)
			if err != nil {
				return nil, err
			}
			result[i] = val
		}

	}
	return result, nil
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
		case "uint256", "int256", "uint", "int":
			typ := big.NewInt(-1)
			types[i] = typ
		case "address":
			typ := common.Address{}
			types[i] = &typ
		case "string":
			typ := ""
			types[i] = &typ
		case "bool":
			typ := false
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

	transformer := Transformer{
		params:  params,
		clients: t.clients,
	}

	err = template.Template(&spec, &transformer.spec, params)
	if err != nil {
		return nil, err
	}
	//log.Println(transformer.spec)

	return transformer, nil
}

func (t Transformer) Execute(ctx context.Context, base map[string]interface{}) (map[string]interface{}, error) {
	w3func, err := w3.NewFunc(t.spec.funcDef(), t.spec.retTypes())
	if err != nil {
		return nil, err
	}

	data := t.spec.retTargets()
	//log.Println(reflect.TypeOf(data[0]))

	args, err := t.spec.argValues()
	if err != nil {
		return nil, err
	}

	err = t.clients[t.spec.ChainID].CallCtx(
		ctx,
		eth.CallFunc(w3func, t.spec.Address, args...).Returns(data...),
	)
	if err != nil {
		return nil, err
	}

	result := ""

	baseBytes, err := json.Marshal(base)
	if err != nil {
		return nil, err
	}

	result = string(baseBytes)

	for i, ret := range t.spec.Returns {
		result, err = sjson.Set(result, ret.Name, &(data[i]))
		if err != nil {
			return nil, err
		}

	}

	var resultMap map[string]interface{}

	err = json.Unmarshal([]byte(result), &resultMap)
	if err != nil {
		return nil, err
	}

	//log.Println(resultMap)

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

func (t Transformer) CreditsConsumed() int {
	return 3
}

func (t Transformer) Deadline() time.Duration {
	return deadline
}

func (t Transformer) Validate() error {
	return nil
}
