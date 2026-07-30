package route

import (
	"net/http"

	"abr-postcode/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterLgCodeRoutes(r *gin.Engine, data *service.AddressData) {
	r.GET("/lg_code/:lg_code", func(c *gin.Context) {
		lgCode := c.Param("lg_code")

		if !requireDigits(c, "lg_code", lgCode, lgCodeLength) {
			return
		}

		city, ok := data.Cities[lgCode]
		if !ok {
			notFound(c, "lg_code")
			return
		}
		c.JSON(http.StatusOK, city)
	})
}
