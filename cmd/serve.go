package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"abr-postcode/internal/config"
	"abr-postcode/internal/data"
	"abr-postcode/internal/middleware"
	"abr-postcode/internal/route"
	"abr-postcode/internal/service"
	"abr-postcode/internal/version"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return serve(config.Load())
	},
}

// newRouter builds the HTTP handler: access logging, then recovery, then CORS,
// compression and the routes. Logging wraps recovery so that a request whose
// handler panics is still recorded, with the 500 recovery produced.
func newRouter(idx *service.AddressData, corsAllowOrigins []string, appVersion, dataModified string) *gin.Engine {
	// The middleware rejects a configuration allowing no origin at all, so an
	// empty list is read as every origin rather than left to panic.
	if len(corsAllowOrigins) == 0 {
		corsAllowOrigins = []string{config.DefaultCORSAllowOrigin}
	}

	r := gin.New()

	r.Use(middleware.AccessLog())
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     corsAllowOrigins,
		AllowMethods:     []string{"GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	route.RegisterRoutes(r, idx, appVersion, dataModified)

	return r
}

func serve(cfg *config.Config) error {
	dataDir := cfg.DataDir

	idx, err := service.LoadAddressData(dataDir)
	if err != nil {
		slog.Error("Failed to load address data", "error", err)
		return err
	}

	r := newRouter(idx, cfg.CORSAllowOrigins, version.Version, data.GetLocalModified(dataDir))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "version", version.Version, "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		slog.Error("Server error", "error", err)
		return err
	case <-ctx.Done():
	}

	slog.Info("Shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		return err
	}

	slog.Info("Server stopped")
	return nil
}
