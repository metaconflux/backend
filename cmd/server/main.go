package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lmittmann/w3"
	"github.com/metaconflux/backend/internal/api/users"
	"github.com/metaconflux/backend/internal/api/users/jwtmaker"
	"github.com/metaconflux/backend/internal/api/users/repository/sqliterepo"
	"github.com/metaconflux/backend/internal/api/v1alpha"
	cache "github.com/metaconflux/backend/internal/cache/ipfs"
	"github.com/metaconflux/backend/internal/hooks"
	"github.com/metaconflux/backend/internal/hooks/api"
	sqliteresolver "github.com/metaconflux/backend/internal/resolver/sqlite"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/metaconflux/backend/internal/transformers"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/contract"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/ipfs"
	"github.com/metaconflux/backend/internal/transformers/core/v1alpha/print"
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

	m := jwtmaker.NewJWTMaker(viper.GetString("auth.secretKey"))

	e := echo.New()
	e.Use(
		middleware.Logger(), // Log everything to stdout
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     []string{"http://localhost:3000"},
			AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
			AllowCredentials: true,
		}),
		middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Skipper: func(c echo.Context) bool {
				if strings.Contains(c.Request().URL.String(), v1alpha.PUBLIC_GROUP) {
					return true
				}
				if strings.Contains(c.Request().URL.String(), users.PUBLIC_GROUP) {
					return true
				}
				return false
			},
			Validator: func(auth string, c echo.Context) (bool, error) {
				claims, err := m.Verify(auth)
				if err != nil {
					return false, err
				}

				c.Set("user", claims)

				return true, nil
			},
		}),
	)

	e.Use()

	e.Pre(middleware.AddTrailingSlash())
	g := e.Group("/api")

	///Transformers

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

	pritnT := print.NewTransformer()

	err = tm.Register(print.GVK, pritnT.WithSpec)
	if err != nil {
		log.Fatal(err)
	}

	err = tm.Register(ipfs.GVK, ipfsT.WithSpec)
	if err != nil {
		log.Fatal(err)
	}

	err = tm.Register(contract.GVK, contractT.WithSpec)
	if err != nil {
		log.Fatal(err)
	}

	///Hooks
	hm := hooks.NewHooksManager()

	hookApi := api.NewHook()

	err = hm.Register(api.TYPE, hookApi)
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(sqlite.Open("./gorm.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	repository, err := sqliterepo.NewSqliteRepository(db)
	if err != nil {
		log.Fatal(err)
	}
	err = repository.Migrate()
	if err != nil {
		log.Fatal(err)
	}

	//r := memory.NewResolver()
	//r, err := file.NewResolver("./resolver.json")
	r, err := sqliteresolver.NewResolver(db)
	if err != nil {
		log.Fatal(err)
	}
	c := cache.NewIPFSCache(url, shell)
	a := v1alpha.NewAPI(c, r, tm, repository, hm)
	a.Register(g)

	u := users.NewUserAPI(m, c, r, repository)
	u.Register(g)

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
