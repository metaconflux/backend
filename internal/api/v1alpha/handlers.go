package v1alpha

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/metaconflux/backend/internal/cache"
	"github.com/metaconflux/backend/internal/resolver"
	"github.com/metaconflux/backend/internal/transformers"
)

type API struct {
	cache        cache.ICache
	resolver     resolver.IResolver
	transformers *transformers.Transformers
}

func NewAPI(cache cache.ICache, resolver resolver.IResolver, transformers *transformers.Transformers) API {
	return API{
		cache:        cache,
		resolver:     resolver,
		transformers: transformers,
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

	err = a.resolver.Set(data.Contract, id, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(200, MetadataResult{Url: fmt.Sprintf("/api/v1alpha/metadata/%s/", data.Contract)})
}

func (a API) Get(c echo.Context) error {
	contract := c.Param("contract")
	tokenId := c.Param("tokenId")

	cacheId := fmt.Sprintf("%s/%s", contract, tokenId)
	log.Printf("Trying cache for %s", cacheId)

	cacheKey, err := a.resolver.Get(cacheId)
	if err == nil {
		var data interface{}
		err = a.cache.Get(cacheKey, &data)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		log.Printf("Using cache for %s (%s)", cacheId, cacheKey)
		return c.JSON(http.StatusOK, data)
	} else {
		if err != resolver.ErrNotFound {
			return c.JSON(http.StatusBadRequest, err)
		}
	}

	//var result map[string]interface{}

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
	params["contract"] = contract

	result, err := a.transformers.Execute(metadata.Transformers, params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	id, err := a.cache.Push(result)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	err = a.resolver.Set(cacheId, id, 1)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, result)
}
