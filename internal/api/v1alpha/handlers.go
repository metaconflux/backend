package v1alpha

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/metaconflux/backend/internal/api/users/repository"
	"github.com/metaconflux/backend/internal/cache"
	"github.com/metaconflux/backend/internal/hooks"
	"github.com/metaconflux/backend/internal/resolver"
	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/sirupsen/logrus"
)

const (
	VERSION             = "v1alpha"
	PUBLIC_GROUP        = "metadata"
	AUTHENTICATED_GROUP = "manifest"
)

type API struct {
	cache        cache.ICache
	resolver     resolver.IResolver
	repository   repository.UserRepository
	transformers *transformers.Transformers
	hooks        hooks.HookManager
}

func NewAPI(
	cache cache.ICache,
	resolver resolver.IResolver,
	transformers *transformers.Transformers,
	repository repository.UserRepository,
	hooks hooks.HookManager,
) API {
	return API{
		cache:        cache,
		resolver:     resolver,
		transformers: transformers,
		repository:   repository,
		hooks:        hooks,
	}
}

func (a API) Register(g *echo.Group) {
	ag := g.Group(fmt.Sprintf("/%s/%s", VERSION, AUTHENTICATED_GROUP))
	ag.POST("/", a.Create)
	ag.GET("/", a.List)
	ag.PUT("/:chainId/:contract/", a.Update)
	ag.GET("/:chainId/:contract/", a.Get)
	ag.GET("/:chainId/:contract/refresh/:tokenId/", a.Refresh)

	publicG := g.Group(fmt.Sprintf("/%s/%s", VERSION, PUBLIC_GROUP))
	publicG.GET("/:chainId/:contract/:tokenId/", a.GetMetadata)
}

func (a API) Create(c echo.Context) error {
	var data Manifest
	err := c.Bind(&data)
	normalizedContract := strings.ToLower(data.Contract)

	if !data.ValidVersion(VERSION) {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Invalid Resource Version")))
	}

	chainId := fmt.Sprintf("%d", data.ChainID)
	manifest, _, err := a.getMetadata(a.formatChainContractKey(chainId, normalizedContract))
	if err != nil {
		if err != resolver.ErrNotFound && err != resolver.ErrLifetime {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}
	}

	log.Println(manifest)

	if len(manifest.Owner) > 0 && len(manifest.Contract) > 0 {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Resource Already Exists")))
	}

	user, err := a.getUser(c)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
	}

	data.Owner = user.Subject

	id, err := a.cache.Push(data)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	err = a.resolver.Set(a.formatChainContractKey(chainId, normalizedContract), id, 0)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	if len(data.Config.Alias) > 0 {
		aliasKey := a.formatChainContractKey(chainId, data.Config.Alias)
		_, err := a.resolver.Get(aliasKey)
		if err != nil && err != resolver.ErrNotFound {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}

		err = a.resolver.Set(aliasKey, id, 0)
		if err != nil {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}
	}

	um, err := a.repository.GetByAddress(c.Request().Context(), user.Subject)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	mm := repository.ManifestModel{
		Address: data.Contract,
		ChainId: data.ChainID,
		User:    um,
	}
	err = a.repository.CreateManifest(c.Request().Context(), mm)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusCreated, MetadataResult{Url: fmt.Sprintf("/api/v1alpha/metadata/%s/", normalizedContract)})
}

func (a API) Update(c echo.Context) error {
	var data Manifest
	err := c.Bind(&data)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
	}
	contract := strings.ToLower(c.Param("contract"))
	chainId := c.Param("chainId")
	normalizedContract := strings.ToLower(data.Contract)

	if !data.ValidVersion(VERSION) {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Invalid Resource Version")))
	}

	if contract != normalizedContract {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("Contract parameter does not match the payload")))
	}

	if chainId != fmt.Sprintf("%d", data.ChainID) {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, fmt.Errorf("ChainId parameter does not match the payload")))
	}

	manifest, _, err := a.getMetadata(a.formatChainContractKey(chainId, normalizedContract))
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	user, err := a.ensureOwner(c, manifest)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusUnauthorized, err))
	}

	um, err := a.repository.GetByAddress(c.Request().Context(), user.Subject)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	data.Owner = user.Subject

	id, err := a.cache.Push(data)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	if len(data.Config.Alias) > 0 || data.Config.Alias != manifest.Config.Alias {
		err := a.validateTier(data, um.TierID)
		if err != nil {
			return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
		}

		aliasKey := a.formatChainContractKey(chainId, data.Config.Alias)

		if data.Config.Alias != manifest.Config.Alias && len(data.Config.Alias) > 0 {
			_, err := a.resolver.Get(aliasKey)
			if err == nil {
				return c.JSON(utils.NewApiError(http.StatusInternalServerError, fmt.Errorf("Failed to update alias - already used")))
			} else if err != resolver.ErrNotFound {
				return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
			}
		}

		err = a.resolver.Set(aliasKey, id, 0)
		if err != nil {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}

		if data.Config.Alias != manifest.Config.Alias {
			log.Println("Deleting cache key")
			oldKey := a.formatChainContractKey(chainId, manifest.Config.Alias)
			err = a.resolver.Delete(oldKey)
			if err != nil {
				return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
			}
		}
	}

	err = a.resolver.Set(a.formatChainContractKey(chainId, normalizedContract), id, 0)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusOK, MetadataResult{Url: fmt.Sprintf("/api/v1alpha/metadata/%s/", normalizedContract)})
}

func (a API) Get(c echo.Context) error {
	contract := strings.ToLower(c.Param("contract"))
	chainId := c.Param("chainId")

	metadata, _, err := a.getMetadata(a.formatChainContractKey(chainId, contract))
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	_, err = a.ensureOwner(c, metadata)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusUnauthorized, err))
	}

	return c.JSON(http.StatusOK, metadata)
}

func (a API) List(c echo.Context) error {
	user, err := a.getUser(c)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
	}

	um, err := a.repository.GetByAddress(c.Request().Context(), user.Subject)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	manifests, err := a.repository.GetManifests(c.Request().Context(), um.ID)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	result := make([]ManifestList, 0)
	for _, m := range manifests {
		item := ManifestList{
			Address: m.Address,
			ChainId: m.ChainId,
		}
		log.Println(a.formatChainContractKey(fmt.Sprintf("%d", m.ChainId), strings.ToLower(m.Address)))
		tmpManifest, _, err := a.getMetadata(a.formatChainContractKey(fmt.Sprintf("%d", m.ChainId), strings.ToLower(m.Address)))
		if err != nil {
			log.Println(err)
			continue
		}

		item.Alias = tmpManifest.Config.Alias

		result = append(result, item)

	}

	return c.JSON(http.StatusOK, result)
}

func (a API) Refresh(c echo.Context) error {
	contract := strings.ToLower(c.Param("contract"))
	chainId := c.Param("chainId")
	tokenId := c.Param("tokenId")

	manifest, _, err := a.getMetadata(a.formatChainContractKey(chainId, contract))
	if err != nil {
		logrus.Errorf("Failed to load manifest: %e", err)
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	_, err = a.ensureOwner(c, manifest)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusUnauthorized, err))
	}

	result, err := a.generate(tokenId, chainId, contract)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusOK, result)
}

func (a API) GetMetadata(c echo.Context) error {
	contract := strings.ToLower(c.Param("contract"))
	chainId := c.Param("chainId")
	tokenId := c.Param("tokenId")

	cacheId := fmt.Sprintf("%s/%s", contract, tokenId) //FIXME
	log.Printf("Trying cache for %s", cacheId)

	cacheKey, err := a.resolver.Get(cacheId)
	if err == nil {
		log.Printf("Got cache key %s", cacheKey)
		var data interface{}
		err = a.cache.Get(cacheKey, &data)
		if err != nil {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}

		log.Printf("Using cache for %s (%s)", cacheId, cacheKey)
		return c.JSON(http.StatusOK, data)
	} else {
		if err == resolver.ErrLifetime {
			manifest, _, err := a.getMetadata(a.formatChainContractKey(chainId, contract))
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

	result, err := a.generate(tokenId, chainId, contract)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusOK, result)
}

func (a API) getMetadata(manifestKey string) (Manifest, string, error) {
	var metadata Manifest

	key, err := a.resolver.Get(manifestKey)
	if err != nil {
		return metadata, "", err
	}

	err = a.cache.Get(key, &metadata)
	if err != nil {
		return metadata, "", err
	}

	return metadata, key, err
}

func (a API) getUser(c echo.Context) (*jwt.StandardClaims, error) {
	user, ok := c.Get("user").(*jwt.StandardClaims)
	if !ok {
		return nil, fmt.Errorf("Failed to load user data")
	}
	return user, nil
}

func (a API) ensureOwner(c echo.Context, metadata Manifest) (*jwt.StandardClaims, error) {
	user, err := a.getUser(c)
	if err != nil {
		return nil, err
	}

	log.Println(user.Subject)
	if metadata.Owner != user.Subject {
		return nil, fmt.Errorf("Failed to load:User is not owner of the resource")
	}

	return user, nil

}

func (a API) formatChainContractKey(chainId string, contract string) string {
	return fmt.Sprintf("manifest#%s#%s", chainId, contract)
}

func (a API) validateTier(manifest Manifest, tierId uint) error {
	var errs []string

	switch tierId {
	case 0:
		if len(manifest.Config.Alias) > 0 {
			errs = append(errs, "Alias is not available in your tier")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "."))
	}

	return nil
}

func (a API) generate(tokenId string, chainId string, contract string) (map[string]interface{}, error) {
	cacheId := fmt.Sprintf("%s/%s", contract, tokenId)

	manifest, manifestCID, err := a.getMetadata(a.formatChainContractKey(chainId, contract))
	if err != nil {
		return nil, err
	}

	params := make(map[string]interface{})
	params["id"] = tokenId
	params["contract"] = contract
	params["manifestCID"] = manifestCID

	log.Println("Credits:", a.transformers.CalculateCredits(manifest.Transformers))

	result, err := a.transformers.Execute(manifest.Transformers, params)
	if err != nil {
		logrus.Errorf("Failed while executing transformers: %s", err)
		return nil, err
	}

	id, err := a.cache.Push(result)
	if err != nil {
		return nil, err
	}

	err = a.resolver.Set(cacheId, id, manifest.Config.RefreshAfter.ToMinute())
	if err != nil {
		return nil, err
	}

	for _, h := range manifest.Hooks {
		log.Println("Hooking", h)
		hook, err := a.hooks.Get(h.Type)
		if err != nil {
			log.Println("Hook", h.Type, " failed", err)
			break
		}

		hook, err = hook.WithSpec(h.Spec, params)
		if err != nil {
			log.Println("Hook", h.Type, " failed", err)
			break
		}

		go func() {
			err = hook.Execute()
			if err != nil {
				logrus.Errorf("Hook %s failed: %s", h.Type, err)
			}
		}()
	}

	return result, nil
}
