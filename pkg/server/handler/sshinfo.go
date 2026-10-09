package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SSHInfo serves GET /api/ssh/info. A nil lookup, or a nil body, is the
// disabled listener: {"enabled": false}.
func SSHInfo(lookup func() any) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := any(gin.H{"enabled": false})
		if lookup != nil {
			if got := lookup(); got != nil {
				body = got
			}
		}
		c.JSON(http.StatusOK, body)
	}
}
