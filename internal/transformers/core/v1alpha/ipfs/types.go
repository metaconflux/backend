package ipfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/metaconflux/backend/internal/gvk"
	"github.com/metaconflux/backend/internal/template"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/sirupsen/logrus"
)

var GVK = gvk.NewGroupVersionKind(
	"core",
	"v1alpha",
	"ipfs",
)

var deadline = 3 * time.Second

var _ transformers.ITransformer = (*Transformer)(nil)

type Transformer struct {
	spec        SpecSchema
	params      map[string]interface{}
	data        map[string]interface{}
	initialized bool
	ipfsClient  *shell.Shell
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

	transformer := Transformer{
		params:      params,
		ipfsClient:  t.ipfsClient,
		initialized: true,
	}

	err = template.Template(&spec, &transformer.spec, params)
	if err != nil {
		return nil, err
	}

	return transformer, nil
}

func (t Transformer) Execute(ctx context.Context, base map[string]interface{}) (map[string]interface{}, error) {
	split := strings.TrimPrefix(t.spec.Url, "ipfs://")

	path, err := utils.Template(split, t.params)
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
	return t.params
}

func (t Transformer) Result() interface{} {
	return nil
}

func (t Transformer) CreditsConsumed() int {
	return 1
}

func (s Transformer) Copy(spec SpecSchema) error {
	cid, err := getCID(spec.Url)
	if err != nil {
		return err
	}
	return s.ipfsClient.FilesCp(context.Background(), fmt.Sprintf("/ipfs/%s", cid), fmt.Sprintf("/%s", cid))
}

type SpecSchema struct {
	Url string `json:"url" template:""`
}

func getCID(url string) (string, error) {
	re := regexp.MustCompile("ipfs://([^/]+)/?(.*)")
	match := re.FindAllStringSubmatch(url, -1)
	logrus.Infoln(match)
	if len(match) == 0 {
		return "", fmt.Errorf("Failed to parse IPFS URL")
	}

	if len(match[0]) < 2 {
		return "", fmt.Errorf("Failed to parse IPFS URL")
	}
	return match[0][1], nil
}

func (t Transformer) Deadline() time.Duration {
	return deadline
}

func (t Transformer) Validate() error {
	if !t.initialized {
		return fmt.Errorf("Not initialized")
	}

	c, err := getCID(t.spec.Url)
	if err != nil {
		return err
	}

	_, err = cid.Decode(c)
	if err != nil {
		return err
	}

	return nil
}
