package contract

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/manifoldco/promptui"
	"github.com/metaconflux/backend/internal/chains"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
)

type Input struct {
	Name string
	Type string
}

type Output struct {
	Name string
	Type string
}

type AbiFunc struct {
	Inputs          []Input
	Outputs         []Output
	Name            string
	StateMutability string
}

func NewSpecFromPrompt() (transformers.BaseTransformer, error) {
	var base transformers.BaseTransformer
	prompt := promptui.Prompt{
		Label:   "Contract Address",
		Default: utils.ZERO_ADDR,
	}

	addr, err := prompt.Run()
	if err != nil {
		return base, err
	}

	promptS := chains.PromptSelect
	cid, _, err := promptS.Run()
	if err != nil {
		return base, err
	}

	abiMap, err := chains.GetAbi(cid, addr, "3BM3TVAFFTKECA2I6NIZU8CXRIG5DH51AR")
	if err != nil {
		return base, err
	}

	var abi []AbiFunc
	utils.Remarshal(abiMap, &abi)

	funcNames := []string{}

	for _, item := range abi {
		if item.StateMutability == "view" || item.StateMutability == "pure" {
			funcNames = append(funcNames, item.Name)
		}
	}

	funcSearcher := func(input string, index int) bool {
		fn := strings.ToLower(funcNames[index])
		input = strings.ToLower(input)

		return strings.Contains(fn, input)
	}

	promptS = promptui.Select{
		Label:    "Function",
		Items:    funcNames,
		Searcher: funcSearcher,
	}

	_, funcName, err := promptS.Run()
	if err != nil {
		return base, err
	}

	var abiFunc AbiFunc

	for _, af := range abi {
		if af.Name == funcName {
			abiFunc = af
			break
		}
	}
	utils.JsonPretty(abiFunc)

	spec := SpecSchema{
		Address:  common.HexToAddress(addr),
		ChainID:  uint64(chains.GetChainId(cid)),
		Function: funcName,
		Args:     make([]Arg, len(abiFunc.Inputs)),
		Returns:  make([]Ret, len(abiFunc.Outputs)),
	}

	for aid, arg := range abiFunc.Inputs {
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("Argument %s (%s)", arg.Name, arg.Type),
		}

		spec.Args[aid].Type = arg.Type
		spec.Args[aid].Value, err = prompt.Run()
	}

	for rid, arg := range abiFunc.Outputs {
		name := arg.Name
		if len(name) == 0 {
			name = strconv.Itoa(rid)
		}
		prompt := promptui.Prompt{
			Label: fmt.Sprintf("Return %s (%s)", name, arg.Type),
		}

		spec.Returns[rid].Type = arg.Type
		spec.Returns[rid].Name, err = prompt.Run()
	}

	base = transformers.BaseTransformer{
		GroupVersionKind: GVK,
		Spec:             spec,
	}

	return base, nil

}
