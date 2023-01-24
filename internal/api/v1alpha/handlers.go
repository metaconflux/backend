package v1alpha

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/metaconflux/backend/internal/cache"
	"github.com/metaconflux/backend/internal/resolver"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
)

const (
	VERSION             = "v1alpha"
	PUBLIC_GROUP        = "metadata"
	AUTHENTICATED_GROUP = "manifest"
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
	ag := g.Group(fmt.Sprintf("/%s/%s", VERSION, AUTHENTICATED_GROUP))
	ag.POST("/", a.Create)
	ag.PUT("/:contract/", a.Update)
	ag.GET("/:contract/", a.Get)

	publicG := g.Group(fmt.Sprintf("/%s/%s", VERSION, PUBLIC_GROUP))
	publicG.GET("/:contract/:tokenId/", a.GetMetadata)
}

func (a API) Create(c echo.Context) error {
	var data Manifest
	err := c.Bind(&data)
	normalizedContract := strings.ToLower(data.Contract)

	if !data.ValidVersion(VERSION) {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Invalid Resource Version")))
	}

	manifest, err := a.getByAddress(normalizedContract)
	if err != nil {
		if err != resolver.ErrNotFound && err != resolver.ErrLifetime {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}
	}

	log.Println(manifest)

	if len(manifest.Owner) > 0 && len(manifest.Contract) > 0 {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Resource Already Exists")))
	}

	user := c.Get("user").(*jwt.StandardClaims)
	data.Owner = user.Subject

	id, err := a.cache.Push(data)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	err = a.resolver.Set(normalizedContract, id, 0)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusCreated, MetadataResult{Url: fmt.Sprintf("/api/v1alpha/metadata/%s/", normalizedContract)})
}

func (a API) Update(c echo.Context) error {
	var data Manifest
	err := c.Bind(&data)
	contract := strings.ToLower(c.Param("contract"))
	normalizedContract := strings.ToLower(data.Contract)

	if !data.ValidVersion(VERSION) {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Invalid Resource Version")))
	}

	if contract != normalizedContract {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Contract parameter does not match the payload")))
	}

	manifest, err := a.getByAddress(normalizedContract)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	user, err := a.ensureOwner(c, manifest)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusUnauthorized, err))
	}

	data.Owner = user.Subject

	id, err := a.cache.Push(data)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	err = a.resolver.Set(normalizedContract, id, 0)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusOK, MetadataResult{Url: fmt.Sprintf("/api/v1alpha/metadata/%s/", normalizedContract)})
}

func (a API) Get(c echo.Context) error {
	contract := strings.ToLower(c.Param("contract"))

	metadata, err := a.getByAddress(contract)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	_, err = a.ensureOwner(c, metadata)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusUnauthorized, err))
	}

	return c.JSON(http.StatusOK, metadata)
}

func (a API) GetMetadata(c echo.Context) error {
	contract := strings.ToLower(c.Param("contract"))
	tokenId := c.Param("tokenId")

	cacheId := fmt.Sprintf("%s/%s", contract, tokenId)
	log.Printf("Trying cache for %s", cacheId)

	cacheKey, err := a.resolver.Get(cacheId)
	if err == nil {
		var data interface{}
		err = a.cache.Get(cacheKey, &data)
		if err != nil {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}

		log.Printf("Using cache for %s (%s)", cacheId, cacheKey)
		return c.JSON(http.StatusOK, data)
	} else {
		if err == resolver.ErrLifetime {
			manifest, err := a.getByAddress(contract)
			if err != nil {
				return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
			}

			if manifest.Config.Freeze {
				err = a.resolver.Set(cacheId, cacheKey, manifest.Config.RefreshAfter.ToMinute())
				if err != nil {
					return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
				}

				var data interface{}
				err = a.cache.Get(cacheKey, &data)
				if err != nil {
					return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
				}

				log.Printf("Frozen: renewing and using cache for %s (%s)", cacheId, cacheKey)
				return c.JSON(http.StatusOK, data)
			}
		} else if err != resolver.ErrNotFound {
			return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
		}

	}

	//var result map[string]interface{}

	if contract == "" {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Contract address parameter empty")))
	}

	manifest, err := a.getByAddress(contract)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	params := make(map[string]interface{})
	params["id"] = tokenId
	params["contract"] = contract

	result, err := a.transformers.Execute(manifest.Transformers, params)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	id, err := a.cache.Push(result)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	err = a.resolver.Set(cacheId, id, manifest.Config.RefreshAfter.ToMinute())
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusOK, result)
}

func (a API) getByAddress(contract string) (Manifest, error) {
	var metadata Manifest

	key, err := a.resolver.Get(contract)
	if err != nil {
		return metadata, err
	}

	err = a.cache.Get(key, &metadata)
	if err != nil {
		return metadata, err
	}

	return metadata, err
}

func (a API) ensureOwner(c echo.Context, metadata Manifest) (*jwt.StandardClaims, error) {
	user, ok := c.Get("user").(*jwt.StandardClaims)
	if !ok {
		return nil, fmt.Errorf("Failed to load user data")
	}

	log.Println(user.Subject)
	if metadata.Owner != user.Subject {
		return nil, fmt.Errorf("Failed to load:User is not owner of the resource")
	}

	return user, nil

}
