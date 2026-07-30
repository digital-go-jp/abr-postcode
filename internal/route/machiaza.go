package route

import (
	"net/http"

	"abr-postcode/internal/model"
	"abr-postcode/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterMachiazaRoutes(r *gin.Engine, data *service.AddressData) {
	r.GET("/machiaza/:lg_code/:machiaza_id", func(c *gin.Context) {
		lgCode := c.Param("lg_code")
		machiazaID := c.Param("machiaza_id")

		if !requireDigits(c, "lg_code", lgCode, lgCodeLength) {
			return
		}

		if !requireDigits(c, "machiaza_id", machiazaID, machiazaIDLength) {
			return
		}

		address, ok := data.Address(lgCode, machiazaID)
		if !ok {
			notFound(c, "machiaza")
			return
		}

		postCodes := data.PostCodes(lgCode, machiazaID)
		if postCodes == nil {
			postCodes = []string{}
		}

		c.JSON(http.StatusOK, model.MachiazaResponse{
			Address:   address,
			PostCodes: postCodes,
		})
	})
}
