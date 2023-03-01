package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/lmittmann/w3"
	"github.com/metaconflux/backend/internal/api/v1alpha"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/contract"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/ipfs"
	"github.com/metaconflux/backend/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Need path to manifest")
	}

	tm, _ := transformers.NewTransformerManager()

	url := "http://localhost:5001"
	shell := shell.NewShellWithClient(url, &http.Client{})
	ipfsT := ipfs.NewTransformer(shell)

	clients := make(map[uint64]*w3.Client)

	var err error
	clients[80001] = w3.MustDial("https://polygon-testnet.public.blastapi.io")
	if err != nil {
		log.Fatal(err)
	}
	defer clients[80001].Close()

	constractT := contract.NewTransformer(clients)

	err = tm.Register(ipfs.GVK, ipfsT.WithSpec, ipfs.NewSpecFromPrompt)
	if err != nil {
		log.Fatal(err)
	}

	err = tm.Register(contract.GVK, constractT.WithSpec, contract.NewSpecFromPrompt)
	if err != nil {
		log.Fatal(err)
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var manifest v1alpha.Manifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		log.Fatal(err)
	}

	result, err := tm.Execute(manifest.Transformers, map[string]interface{}{"id": 1})
	if err != nil {
		log.Fatal(err)
	}

	err = utils.JsonPretty(result)
	if err != nil {
		log.Fatal(err)
	}
}
