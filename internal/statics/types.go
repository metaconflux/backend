package statics

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

type IStatics interface {
	Get(spec SpecSchema, params map[string]interface{}) (map[string]interface{}, error)
	Copy(spec SpecSchema) error
}

type Statics struct {
	ipfsClient *shell.Shell
}

func NewStatics(shell *shell.Shell) *Statics {
	return &Statics{
		ipfsClient: shell,
	}
}

func (s Statics) Get(spec SpecSchema, params map[string]interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}

	split := strings.Split(spec.Url, ":")
	switch split[0] {
	case "ipfs":
		r, err := s.ipfsClient.Cat(path.Join(split[1][2:], params["id"].(string)))
		if err != nil {
			return nil, err
		}

		defer r.Close()
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &result)
		log.Println(string(data))
		log.Println(result)
	default:
		log.Println(split)
		return nil, fmt.Errorf("Unknown protocol in URL")
	}

	return result, nil
}

func (s Statics) Copy(spec SpecSchema) error {
	cid := getCID(spec.Url)
	return s.ipfsClient.FilesCp(context.Background(), fmt.Sprintf("/ipfs/%s", cid), fmt.Sprintf("/%s", cid))
}

type SpecSchema struct {
	Url string `json:"url"`
}

func getCID(url string) string {
	return strings.TrimPrefix(url, "ipfs://")
}
