package ipfs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/metaconflux/backend/internal/cache"
)

type IPFSCache struct {
	cache.ICache
	url    string
	client *shell.Shell
}

func NewIPFSCache(url string, shell *shell.Shell) *IPFSCache {
	return &IPFSCache{
		url:    url,
		client: shell,
	}
}

func (c IPFSCache) Get(id string, target interface{}) error {
	reader, err := c.client.Cat(id)
	if err != nil {
		return err
	}

	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &target)
	if err != nil {
		return err
	}

	return nil
}

func (c IPFSCache) Push(data interface{}) (string, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	r := bytes.NewReader(b)

	cid, err := c.client.Add(r)
	if err != nil {
		return "", err
	}

	return cid, nil
}
