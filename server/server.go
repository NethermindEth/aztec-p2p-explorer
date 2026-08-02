// package server implements the HTTP endpoints for Aztec P2P Explorer API.
//
//nolint:misspell
package server

//	@title			Aztec P2P Explorer API
//	@version		1
//  @BasePath 		/api
//	@description	This is the aztec P2P Explorer API server.
//	@contact.name	aztec-p2p-explorer
//	@contact.url	http://github.com/NethermindEth/aztec-p2p-explorer

//	@host	localhost:8080
//	@scheme	http

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/NethermindEth/aztec-p2p-explorer/core"
	coretypes "github.com/NethermindEth/aztec-p2p-explorer/core/types"
	_ "github.com/NethermindEth/aztec-p2p-explorer/docs"
	"github.com/NethermindEth/aztec-p2p-explorer/internal/turnstile"
	"github.com/NethermindEth/aztec-p2p-explorer/repo"
)

const (
	// DefaultCacheSize is the default size for the LRU caches
	DefaultCacheSize = 50

	// DefaultPrecision is the default precision for number rounding
	DefaultPrecision = 10
)

type Server struct {
	e           *echo.Echo
	bindAddress string
}

type Handler struct {
	Repo                    *repo.PeerRepository
	PeersResponseCache      *lru.Cache[uint64, PeersResponse]
	PeersMapResponseCache   *lru.Cache[uint64, PeersMapResponse]
	AnalyticsResponseCache  *lru.Cache[uint64, AnalyticsResponse]
	CountriesResponseCache  *lru.Cache[uint64, CountryCountResponse]
	CitiesResponseCache     *lru.Cache[uint64, CityCountResponse]
	ContinentsResponseCache *lru.Cache[uint64, ContinentCountResponse]
	UseIndexTable           bool
	AztecFeeder             *core.AztecFeederClient
	Metrics                 *BusinessMetrics
}

// GetFilterCCV returns the current CCV from the reference node's ENR
// Returns empty string if AztecFeeder is not configured or CCV is not available
func (h *Handler) GetFilterCCV() string {
	if h.AztecFeeder == nil {
		return ""
	}
	return h.AztecFeeder.GetReferenceCCV()
}

func createCache[T any](logger *slog.Logger, cacheName string) *lru.Cache[uint64, T] {
	cache, err := lru.New[uint64, T](DefaultCacheSize)
	if err != nil {
		logger.Error("failed to create "+cacheName+" cache", "error", err)
		return nil
	}
	return cache
}

func setupCacheCleanupAndMetrics(
	onCrawlProcessedChan chan *coretypes.Crawl,
	logger *slog.Logger,
	caches []interface{ Purge() },
	metrics *BusinessMetrics,
) {
	go func() {
		for crawl := range onCrawlProcessedChan {
			logger.Info("Purging caches and updating metrics")
			for _, cache := range caches {
				if cache != nil {
					cache.Purge()
				}
			}
			// Update metrics after each crawl completion
			metrics.OnCrawlSuccess(context.Background(), crawl)
		}
	}()
}

// privateAPIAuth guards /api/private with BasicAuth. Credentials come from
// the environment; without them the group rejects everything rather than
// falling back to a built-in default.
func privateAPIAuth(logger *slog.Logger) echo.MiddlewareFunc {
	user := os.Getenv("PRIVATE_API_USERNAME")
	pass := os.Getenv("PRIVATE_API_PASSWORD")
	if user == "" || pass == "" {
		logger.Warn("PRIVATE_API_USERNAME/PRIVATE_API_PASSWORD not set; /api/private is disabled")
	}

	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if user == "" || pass == "" {
			return false, nil
		}
		// Constant time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(user)) == 1 &&
			subtle.ConstantTimeCompare([]byte(password), []byte(pass)) == 1 {
			return true, nil
		}
		return false, nil
	})
}

func New(
	bindAddress string,
	logger *slog.Logger,
	r *repo.PeerRepository,
	onCrawlProcessedChan chan *coretypes.Crawl,
	useIndexTable bool,
	aztecFeeder *core.AztecFeederClient,
) *Server {
	e := echo.New()
	e.HideBanner = true

	peersResponseCache := createCache[PeersResponse](logger, "peers response")
	analyticsResponseCache := createCache[AnalyticsResponse](logger, "analytics response")
	peersMapResponseCache := createCache[PeersMapResponse](logger, "peers map response")
	countriesResponseCache := createCache[CountryCountResponse](logger, "countries response")
	citiesResponseCache := createCache[CityCountResponse](logger, "cities response")
	continentsResponseCache := createCache[ContinentCountResponse](logger, "continents response")

	caches := []interface{ Purge() }{
		peersResponseCache,
		analyticsResponseCache,
		peersMapResponseCache,
		countriesResponseCache,
		citiesResponseCache,
		continentsResponseCache,
	}

	metrics := NewBusinessMetrics(logger, r, aztecFeeder)

	metrics.UpdateMetrics(context.Background())
	metrics.StartAgeUpdater(context.Background())

	setupCacheCleanupAndMetrics(onCrawlProcessedChan, logger, caches, metrics)

	// Middlewares
	e.Use(slogecho.New(logger))
	e.Use(middleware.Recover())

	e.Use(echoprometheus.NewMiddleware("aztec_p2p_explorer"))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))

	e.HTTPErrorHandler = customErrorHandler(e)

	h := &Handler{
		Repo:                    r,
		PeersResponseCache:      peersResponseCache,
		PeersMapResponseCache:   peersMapResponseCache,
		AnalyticsResponseCache:  analyticsResponseCache,
		CountriesResponseCache:  countriesResponseCache,
		CitiesResponseCache:     citiesResponseCache,
		ContinentsResponseCache: continentsResponseCache,
		UseIndexTable:           useIndexTable,
		AztecFeeder:             aztecFeeder,
		Metrics:                 metrics,
	}

	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/health", h.Health)

	api := e.Group("/api")
	privateAPI := e.Group("/api/private")

	// Turnstile protection middleware (configurable via env)
	cfg := turnstile.FromEnv()
	if cfg.Mode != turnstile.ModeDisabled && cfg.SecretKey != "" {
		ts := &turnstile.Middleware{
			Config:   cfg,
			Verifier: turnstile.NewHTTPVerifier(cfg.SecretKey, cfg.Timeout),
			Cache:    turnstile.NewMemoryCache(),
		}
		api.Use(echo.WrapMiddleware(ts.Echo()))
		privateAPI.Use(echo.WrapMiddleware(ts.Echo()))
	}
	privateAPI.Use(privateAPIAuth(logger))

	// SPA fallback handler - serves index.html for non-API routes
	e.GET("/*", func(c echo.Context) error {
		path := c.Request().URL.Path
		// If it's an API route, let it 404 naturally
		if strings.HasPrefix(path, "/api") {
			return echo.ErrNotFound
		}

		// Special handling for root path and index.html - always inject config
		if path == "/" || path == "/index.html" {
			indexFile, err := Frontend().Open("index.html")
			if err != nil {
				return echo.ErrNotFound
			}
			defer indexFile.Close()

			// Read index.html content
			content, err := io.ReadAll(indexFile)
			if err != nil {
				return echo.ErrInternalServerError
			}

			// Inject runtime configuration into HTML
			html := injectRuntimeConfig(string(content), cfg)

			return c.HTML(http.StatusOK, html)
		}

		// Try to serve the file first (for assets)
		file, err := Frontend().Open(path)
		if err == nil {
			defer file.Close()
			// File exists, serve it
			return echo.WrapHandler(http.FileServer(Frontend()))(c)
		}

		// File doesn't exist, serve index.html for SPA routing
		indexFile, err := Frontend().Open("index.html")
		if err != nil {
			return echo.ErrNotFound
		}
		defer indexFile.Close()

		// Read index.html content
		content, err := io.ReadAll(indexFile)
		if err != nil {
			return echo.ErrInternalServerError
		}

		// Inject runtime configuration into HTML
		html := injectRuntimeConfig(string(content), cfg)

		return c.HTML(http.StatusOK, html)
	})

	api.GET("/peers", h.GetPeers)
	api.GET("/peers/map", h.GetPeersMap)
	api.GET("/analytics/peers/history", h.GetPeerHistory)
	api.GET("/analytics/continents", h.GetPeerCountByContinent)
	api.GET("/analytics/countries", h.GetPeerCountByCountry)
	api.GET("/analytics/clients", h.GetPeerCountByAgent)
	api.GET("/analytics", h.GetAnalytics)

	privateAPI.GET("/peers", h.GetPeersPrivate)
	privateAPI.GET("/peers/neighbors", h.GetPeersNeighbors)
	privateAPI.GET("/peers/:id", h.GetPeerByID)
	privateAPI.GET("/peers/:id/neighbors", h.GetPeerNeighbors)
	privateAPI.GET("/analytics", h.GetAnalytics)
	privateAPI.GET("/analytics/peers", h.GetPeerCount)
	privateAPI.GET("/analytics/clients", h.GetPeerCountByAgent)
	privateAPI.GET("/analytics/protocols", h.GetPeerCountByProtocol)
	privateAPI.GET("/analytics/continents", h.GetPeerCountByContinent)
	privateAPI.GET("/analytics/countries", h.GetPeerCountByCountry)
	privateAPI.GET("/analytics/cities", h.GetPeerCountByCity)
	privateAPI.GET("/analytics/networks", h.GetPeerCountByASO)

	// Start server
	return &Server{
		e:           e,
		bindAddress: bindAddress,
	}
}

func (s *Server) Start() error {
	if err := s.e.Start(s.bindAddress); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.e.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

func customErrorHandler(e *echo.Echo) func(error, echo.Context) {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		if myErr, ok := err.(APIError); ok {
			if err := c.JSON(myErr.Status, myErr); err != nil {
				e.Logger.Error(err)
			}
			return
		}
		e.DefaultHTTPErrorHandler(err, c)
	}
}

// injectRuntimeConfig injects runtime configuration into the HTML
func injectRuntimeConfig(html string, turnstileConfig turnstile.Config) string {
	// Get Turnstile site key from environment
	turnstilesiteKey := os.Getenv("TURNSTILE_SITE_KEY")
	if turnstilesiteKey == "" {
		// Try alternate name for compatibility
		turnstilesiteKey = os.Getenv("TURNSTILE_KEY")
	}

	// Only inject config if we have a site key
	if turnstilesiteKey == "" {
		return html
	}

	// Build the config object
	config := map[string]interface{}{
		"turnstile": map[string]interface{}{
			"siteKey": turnstilesiteKey,
			"mode":    string(turnstileConfig.Mode),
		},
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		return html
	}

	// Inject the config script before the closing </head> tag
	configScript := fmt.Sprintf(`<script>window.__APP_CONFIG__=%s;</script>`, configJSON)
	return strings.Replace(html, "</head>", configScript+"</head>", 1)
}
