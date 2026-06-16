package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/service"
)

// UsersHandler ...
type UsersHandler struct {
	Service service.UsersService
}

// GetUser returns the user info
func (handler *UsersHandler) GetUser(c *gin.Context) {

	if userid, ok := c.Get("userid"); ok {
		if user, err := handler.Service.FindByRefID(userid.(string)); err == nil {
			c.JSON(http.StatusOK, user)
			return
		} else {
			c.Error(err)
			return
		}
	}
	c.Error(errors.New("No userid found"))
}

// UpdateUser update the current user information
func (handler *UsersHandler) UpdateUser(c *gin.Context) {

	if userid, ok := c.Get("userid"); ok {

		refID, ok := userid.(string)
		if !ok || refID == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid userid"})
			return
		}

		var updates e.User
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := handler.Service.FindByRefID(refID)
		if err != nil {
			c.Error(err)
			return
		}

		user.Name = updates.Name
		user.Email = updates.Email
		user.Nickname = updates.Nickname
		user.Picture = updates.Picture
		if user.IsNewUser && !updates.IsNewUser {
			user.IsNewUser = false
		}

		if err := handler.Service.Update(refID, user); err == nil {
			c.JSON(http.StatusOK, user)
			return
		} else {
			c.Error(err)
			return
		}
	}
	c.Error(errors.New("No userid found"))
}
