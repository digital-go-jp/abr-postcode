package route

import (
	"abr-postcode/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, data *service.AddressData, version, dataModified string) {
	RegisterHealthRoutes(r, version, dataModified)
	RegisterLgCodeRoutes(r, data)
	RegisterMachiazaRoutes(r, data)
	RegisterPostcodeRoutes(r, data)
}
