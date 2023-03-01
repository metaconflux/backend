package chains

import (
	"github.com/lmittmann/w3"
	"github.com/sirupsen/logrus"
)

type Chain struct {
	Name    string
	ChainId int64
	Rpc     string
	ScanAPI string
}

var Chains []Chain = []Chain{
	{Name: "Mumbai", ChainId: 80001, Rpc: "https://polygon-testnet.public.blastapi.io", ScanAPI: "https://api-testnet.polygonscan.com/"},
}

func GetChainById(chainId int64) Chain {
	for _, c := range Chains {
		if c.ChainId == chainId {
			return c
		}
	}

	return Chain{}
}

type Clients struct {
	clients map[uint64]*w3.Client
}

func NewClients(rpcs map[int64]string) *Clients {
	c := &Clients{}
	c.clients = make(map[uint64]*w3.Client)

	for _, chain := range Chains {
		rpc := chain.Rpc
		if overrideRpc, ok := rpcs[chain.ChainId]; ok {
			rpc = overrideRpc
		}
		c.clients[uint64(chain.ChainId)] = w3.MustDial(rpc)
	}

	return c
}

func (c Clients) Clients() map[uint64]*w3.Client {
	return c.clients
}

func (c Clients) Close() {
	for _, client := range c.clients {
		err := client.Close()
		if err != nil {
			logrus.Error(err)
		}
	}
}
