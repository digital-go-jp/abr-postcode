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

		key := lgCode + machiazaID
		town, ok := data.Towns[key]
		if !ok {
			notFound(c, "machiaza")
			return
		}
		city, ok := data.Cities[town.LgCode]
		if !ok {
			notFound(c, "machiaza")
			return
		}

		// Use pre-built reverse index (O(1) lookup, already sorted)
		postCodes := data.TownToPostCodes[key]
		if postCodes == nil {
			postCodes = []string{}
		}

		c.JSON(http.StatusOK, model.MachiazaResponse{
			Address:   model.NewAddress(town, city),
			PostCodes: postCodes,
		})
	})
}
