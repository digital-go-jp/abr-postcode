package route

import (
	"net/http"

	"abr-postcode/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoutes registers the health check endpoint. The application
// version and ABR data-modified timestamp are exposed as response headers on
// this endpoint only.
func RegisterHealthRoutes(r *gin.Engine, version, dataModified string) {
	r.GET("/health", middleware.Metadata(version, dataModified), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
