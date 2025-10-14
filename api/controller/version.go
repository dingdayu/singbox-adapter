// Package controller defines HTTP controllers.
package controller

import (
	"net/http"

	"github.com/dingdayu/go-project-template/model/entity"

	"github.com/gin-gonic/gin"
)

// Version returns current build version and time.
// @Summary Get API version
// @Produce  json
// @SuccessResponse 200 {object} api.Response
// @Failure 500 {object} api.Response
// @Router /api/v1/version [get]
func Version(c *gin.Context) {
	res := map[string]interface{}{}
	res["code"] = 200
	res["message"] = "success"
	res["data"] = map[string]string{
		"time":    entity.BuildTime,
		"version": entity.BuildVersion,
	}
	c.JSON(http.StatusOK, res)
}
