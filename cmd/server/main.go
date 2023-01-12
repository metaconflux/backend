package main

import (
	"log"
	"net/http"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/metaconflux/backend/internal/api/v1alpha"
	"github.com/metaconflux/backend/internal/cache/ipfs"
	"github.com/metaconflux/backend/internal/resolver/memory"
	"github.com/metaconflux/backend/internal/statics"
)

func main() {
	e := echo.New()
	e.Use(
		middleware.Logger(), // Log everything to stdout
	)
	g := e.Group("/api/v1alpha")

	url := "http://localhost:5001"
	shell := shell.NewShellWithClient(url, &http.Client{})
	c := ipfs.NewIPFSCache(url, shell)
	r := memory.NewResolver()
	s := statics.NewStatics(shell)

	a := v1alpha.NewAPI(c, r, s)
	a.Register(g)

	log.Println(c)

	log.Fatal(e.Start("localhost:8081"))
}
