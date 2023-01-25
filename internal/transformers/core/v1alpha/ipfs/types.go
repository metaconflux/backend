package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/metaconflux/backend/internal/gvk"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
)

var GVK = gvk.NewGroupVersionKind(
	"core",
	"v1alpha",
	"ipfs",
)

func init() {
	var _ transformers.ITransformer = &Transformer{}
}

type Transformer struct {
	transformers.ITransformer
	spec       SpecSchema
	params     map[string]interface{}
	data       map[string]interface{}
	ipfsClient *shell.Shell
}

func NewTransformer(shell *shell.Shell) *Transformer {
	return &Transformer{
		ipfsClient: shell,
	}
}

func (t Transformer) WithSpec(ispec interface{}, params map[string]interface{}) (transformers.ITransformer, error) {
	var spec SpecSchema
	err := utils.Remarshal(ispec, &spec)
	if err != nil {
		return nil, err
	}

	return Transformer{
		spec:       spec,
		params:     params,
		ipfsClient: t.ipfsClient,
	}, nil
}

func (t Transformer) Execute(base map[string]interface{}) (map[string]interface{}, error) {
	split := strings.Split(t.spec.Url, ":")

	path, err := utils.Template(split[1][2:], t.params)
	if err != nil {
		return nil, err
	}

	r, err := t.ipfsClient.Cat(path)
	if err != nil {
		return nil, err
	}

	defer r.Close()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	//t.data = result
	//TODO: Merge with base!!!

	return result, nil
}

func (t Transformer) Status() []transformers.Status {
	return nil
}

func (t Transformer) Params() map[string]interface{} {
	return nil
}

func (s Transformer) Copy(spec SpecSchema) error {
	cid := getCID(spec.Url)
	return s.ipfsClient.FilesCp(context.Background(), fmt.Sprintf("/ipfs/%s", cid), fmt.Sprintf("/%s", cid))
}

type SpecSchema struct {
	Url string `json:"url"`
}

func getCID(url string) string {
	return strings.Split(strings.TrimPrefix(url, "ipfs://"), "/")[0]
}
