package subscribe

import (
	"net/http"

	"github.com/dingdayu/go-project-template/internal/singbox"
	"github.com/dingdayu/go-project-template/internal/upstream"
	"github.com/gin-gonic/gin"
	singjson "github.com/sagernet/sing/common/json"

	"resty.dev/v3"
)

var rc = resty.New()

func Adapter(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.String(http.StatusBadRequest, "missing url")
		return
	}

	up := upstream.ClashVergeSubscriber{}
	ots, err := up.Outboards(c.Request.Context(), rc, url)
	if err != nil {
		c.String(http.StatusInternalServerError, "failed to get outboards: %v", err)
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
