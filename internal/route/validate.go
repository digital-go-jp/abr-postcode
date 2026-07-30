package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Digit lengths of the path parameters the API accepts.
const (
	lgCodeLength     = 6
	machiazaIDLength = 7
	postCodeLength   = 7
)

// requireDigits reports whether value is exactly length ASCII digits. When it
// is not, it answers the request with 400 and the format error naming param.
func requireDigits(c *gin.Context, param, value string, length int) bool {
	if isDigits(value, length) {
		return true
	}
	writeError(c, http.StatusBadRequest, "Invalid "+param+" format")
	return false
}

// notFound answers the request with 404 and names the resource that the API
// holds no data for.
func notFound(c *gin.Context, resource string) {
	writeError(c, http.StatusNotFound, resource+" not found")
}

// writeError answers the request with status and the error envelope every
// failing response shares.
func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// isDigits reports whether s is exactly length ASCII digits. Full-width digits
// do not qualify.
func isDigits(s string, length int) bool {
	if len(s) != length {
		return false
	}
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
