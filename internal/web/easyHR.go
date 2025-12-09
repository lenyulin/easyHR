package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type EasyHRHandler struct {
}

func NewEasyHRHandler() *EasyHRHandler {
	return &EasyHRHandler{}
}

func (h *EasyHRHandler) RegisiterRoutes(server *gin.Engine) {
	ug := server.Group("/easyhr")
	ug.GET("/search", h.SearchByID)
}

func (h *EasyHRHandler) SearchByID(ctx *gin.Context) {
	ctx.String(http.StatusOK, "User Signed Up Succeed")
	return
}
