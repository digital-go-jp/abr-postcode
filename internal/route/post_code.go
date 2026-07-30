package route

import (
	"net/http"

	"abr-postcode/internal/model"
	"abr-postcode/internal/service"

	"github.com/gin-gonic/gin"
)

func RegisterPostcodeRoutes(r *gin.Engine, data *service.AddressData) {
	r.GET("/post_code/:post_code", func(c *gin.Context) {
		postCode := c.Param("post_code")

		if !requireDigits(c, "post_code", postCode, postCodeLength) {
			return
		}

		var res []model.AddressResponse
		for _, mapping := range data.PostCodeMappings[postCode] {
			address, ok := data.Address(mapping.LgCode, mapping.MachiazaID)
			if !ok {
				continue
			}
			res = append(res, model.AddressResponse{
				Address:  address,
				PostCode: mapping.PostCode,
			})
		}
		// An unknown post_code and one whose mappings all fail to join leave
		// the same nothing to return.
		if len(res) == 0 {
			notFound(c, "post_code")
			return
		}
		c.JSON(http.StatusOK, res)
	})
}
