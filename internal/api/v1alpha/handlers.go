package v1alpha

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metaconflux/backend/internal/cache"
	"github.com/metaconflux/backend/internal/resolver"
	"github.com/metaconflux/backend/internal/statics"
	"github.com/metaconflux/backend/internal/utils"
)

type API struct {
	cache    cache.ICache
	resolver resolver.IResolver
	statics  statics.IStatics
}

func NewAPI(cache cache.ICache, resolver resolver.IResolver, statics statics.IStatics) API {
	return API{
		cache:    cache,
		resolver: resolver,
		statics:  statics,
	}
}

func (a API) Register(g *echo.Group) {
	g.POST("/metadata/", a.Create)
	g.GET("/metadata/:contract/:tokenId/", a.Get)
}

func (a API) Create(c echo.Context) error {
	var data MetadataSchema
	err := c.Bind(&data)

	id, err := a.cache.Push(data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = a.resolver.Set(data.Contract, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = a.statics.Copy(data.Static)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(200, MetadataResult{Url: fmt.Sprintf("/api/v1alpha/metadata/%s/", data.Contract)})
}

func (a API) Get(c echo.Context) error {
	contract := c.Param("contract")
	tokenId := c.Param("tokenId")

	var result map[string]interface{}

	if contract == "" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("Contract address parameter empty"))
	}

	key, err := a.resolver.Get(contract)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	log.Println(key)

	var metadata MetadataSchema
	err = a.cache.Get(key, &metadata)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	params := make(map[string]interface{})
	params["id"] = tokenId

	staticData, err := a.statics.Get(metadata.Static, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	d := utils.MergeMaps(result, staticData)

	d["something"] = "more"

	return c.JSON(http.StatusOK, d)
}
