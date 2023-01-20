package main

import (
	"fmt"
	"log"
	"net/http"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lmittmann/w3"
	"github.com/metaconflux/backend/internal/api/v1alpha"
	cache "github.com/metaconflux/backend/internal/cache/ipfs"
	"github.com/metaconflux/backend/internal/resolver/file"
	"github.com/spf13/viper"

	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/contract"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/ipfs"
)

func main() {
	var err error
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	port := viper.GetInt("server.port")
	host := viper.GetString("server.host")

	e := echo.New()
	e.Use(
		middleware.Logger(), // Log everything to stdout
	)
	e.Pre(middleware.AddTrailingSlash())
	g := e.Group("/api/v1alpha")

	tm, _ := transformers.NewTransformerManager()

	url := viper.GetString("ipfs.apiEndpoint")
	projectId := viper.GetString("ipfs.projectId")
	projectSecret := viper.GetString("ipfs.projectSecret")

	httpClient := &http.Client{}

	if projectId != "" && projectSecret != "" {
		httpClient = &http.Client{
			Transport: authTransport{
				RoundTripper:  http.DefaultTransport,
				ProjectId:     projectId,
				ProjectSecret: projectSecret,
			},
		}
	}

	shell := shell.NewShellWithClient(url, httpClient)
	ipfsT := ipfs.NewTransformer(shell)

	clients := make(map[uint64]*w3.Client)

	clients[80001] = w3.MustDial("https://polygon-testnet.public.blastapi.io")
	defer clients[80001].Close()

	contractT := contract.NewTransformer(clients)

	err = tm.Register(ipfs.GVK, ipfsT.WithSpec)
	if err != nil {
		log.Fatal(err)
	}

	err = tm.Register(contract.GVK, contractT.WithSpec)
	if err != nil {
		log.Fatal(err)
	}

	//r := memory.NewResolver()
	r, err := file.NewResolver("./resolver.json")
	if err != nil {
		log.Fatal(err)
	}
	c := cache.NewIPFSCache(url, shell)
	a := v1alpha.NewAPI(c, r, tm)
	a.Register(g)

	log.Fatal(e.Start(fmt.Sprintf("%s:%d", host, port)))
}

// authTransport decorates each request with a basic auth header.
type authTransport struct {
	http.RoundTripper
	ProjectId     string
	ProjectSecret string
}

func (t authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.SetBasicAuth(t.ProjectId, t.ProjectSecret)
	return t.RoundTripper.RoundTrip(r)
}
