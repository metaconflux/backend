package users

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/metaconflux/backend/internal/resolver"
	"github.com/metaconflux/backend/internal/utils"
	"github.com/spruceid/siwe-go"
)

const (
	VERSION      = "v1"
	PUBLIC_GROUP = "auth"
)

const DEFAULT_LIFETIME = 730 * time.Hour

func (a UserApi) Register(g *echo.Group) {
	ug := g.Group(fmt.Sprintf("/auth/%s", VERSION))
	ug.POST("/", a.SignIn)
	ug.GET("/auth/:address/", a.GetAuthText)
}

func (a UserApi) SignIn(c echo.Context) error {
	var signin SignInBody
	err := c.Bind(&signin)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
	}

	msg, err := siwe.ParseMessage(signin.Text)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
	}

	valid, err := msg.ValidNow()
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	if !valid {
		return c.NoContent(http.StatusUnauthorized)
	}

	um, err := a.repository.GetByAddress(c.Request().Context(), string(msg.GetAddress().Hex()))
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	err = a.repository.NewLogin(c.Request().Context(), um.ID, utils.RandStringBytes(8))
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	token, err := a.jwtmaker.Create(um.Address, signin.Lifetime)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	return c.JSON(http.StatusOK, SignInResult{Token: token})
}

func (a UserApi) GetAuthText(c echo.Context) error {
	address := c.Param("address")
	lifetimeStr := c.Param("lifetime")

	lifetime := DEFAULT_LIFETIME
	var err error
	if len(lifetimeStr) != 0 {

		lifetime, err = time.ParseDuration(lifetimeStr)
		if err != resolver.ErrNotFound {
			return c.JSON(utils.NewApiError(http.StatusBadRequest, err))
		}
	}

	um, err := a.repository.GetByAddress(c.Request().Context(), address)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, fmt.Errorf("Failed to load user")))
	}

	/*um.Address = address
	um.Nonce = utils.RandStringBytes(8)
	um.ID = uuid.NewString()*/

	if um.Address == "" {
		um.Address = address
		um.Nonce = utils.RandStringBytes(8)
		um.ID = uuid.NewString()
		err = a.repository.Create(c.Request().Context(), um)
		if err != nil {
			return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
		}
	}

	var options = map[string]interface{}{
		"statement":      "I accept the Terms of service",
		"chainId":        1,
		"issuedAt":       time.Now(),
		"expirationTime": time.Now().Add(lifetime),
		"notBefore":      time.Now(),
	}

	msg, err := siwe.InitMessage("localhost:8081", um.Address, "http://localhost:8081/api/auth/v1/", um.Nonce, options)
	if err != nil {
		return c.JSON(utils.NewApiError(http.StatusInternalServerError, err))
	}

	//text := getAuthText(user.Address, user.Nonce, lifetime)

	return c.JSON(http.StatusOK, AuthTextResult{Text: msg.String(), SignInData: SignInBody{Address: um.Address, Nonce: um.Nonce, Lifetime: lifetime}})
}

func getAuthResolverKey(address string) string {
	return fmt.Sprintf("mcx#auth#%s", address)
}

func getAuthText(address common.Address, nonce uint, lifetime time.Duration) string {
	return fmt.Sprintf("Signing nonce %d\n as an owner of %s\n to authenticate to MCX platform\n for %s", nonce, address, lifetime)
}
