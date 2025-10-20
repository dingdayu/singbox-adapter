package hub

import (
	"net/http"

	"github.com/dingdayu/go-project-template/internal/proxy"
	"github.com/dingdayu/go-project-template/internal/singbox"
	"github.com/dingdayu/go-project-template/internal/token"
	"github.com/dingdayu/go-project-template/internal/upstream"
	"github.com/gin-gonic/gin"
	singjson "github.com/sagernet/sing/common/json"
)

func Subscribe(c *gin.Context) {
	tks := c.Param("token")

	_, err := token.GetToken(tks)
	if err != nil {
		c.String(http.StatusUnauthorized, "invalid token: %v", err)
		return
	}

	ots := proxy.GetOutbounds[upstream.ProxyOutbound]()

	if len(ots) == 0 {
		c.String(http.StatusInternalServerError, "no outbound available")
		return
	}

	opts, err := singbox.OutboundToProfile(ots)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to convert outbounds to profile: %v", err)
		return
	}

	payload, err := singjson.MarshalContext(c.Request.Context(), opts)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to marshal profile: %v", err)
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
}
