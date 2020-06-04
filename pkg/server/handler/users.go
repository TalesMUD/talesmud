package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/service"
)

//UsersHandler ...
type UsersHandler struct {
	Service service.UsersService
}

//GetUser returns the user info
func (handler *UsersHandler) GetUser(c *gin.Context) {

	if userid, ok := c.Get("userid"); ok {
		if user, err := handler.Service.FindByRefID(userid.(string)); err == nil {
			c.JSON(http.StatusOK, user)
		} else {
			c.Error(err)
		}
	}
	c.Error(errors.New("No userid found"))
}

//UpdateUser update the current user information
func (handler *UsersHandler) UpdateUser(c *gin.Context) {

	if userid, ok := c.Get("userid"); ok {

		var user e.User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if user.RefID != userid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Logged in User id does not match attached user"})
			return
		}

		if err := handler.Service.Update(userid.(string), &user); err == nil {
			c.JSON(http.StatusOK, "User updated")
		} else {
			c.Error(err)
			return
		}
	}
	c.Error(errors.New("No userid found"))
}
