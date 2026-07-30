package middleware

import "github.com/gin-gonic/gin"

// Metadata returns a Gin middleware that attaches the application version and
// the ABR data modification timestamp as headers on the routes it is
// registered on. Empty values are omitted so clients can detect "unknown"
// cases.
func Metadata(version, dataModified string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if version != "" {
			c.Header("X-App-Version", version)
		}
		if dataModified != "" {
			c.Header("X-Data-Modified", dataModified)
		}
		c.Next()
	}
}
