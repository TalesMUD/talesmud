package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/service"
)

// UserManagementHandler handles admin user management endpoints.
type UserManagementHandler struct {
	Service service.UsersService
}

// GetAllUsers returns all users (admin only).
func (h *UserManagementHandler) GetAllUsers(c *gin.Context) {
	users, err := h.Service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// roleRequest is the JSON body for role update requests.
type roleRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateUserRole changes a user's role (admin only).
func (h *UserManagementHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: role is required"})
		return
	}

	if err := h.Service.SetRole(userID, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}

// BanUser bans a user by ID (admin only).
func (h *UserManagementHandler) BanUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.Service.BanUser(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User banned"})
}

// UnbanUser removes a ban from a user by ID (admin only).
func (h *UserManagementHandler) UnbanUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.Service.UnbanUser(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unbanned"})
}

// DeleteUser permanently removes a user record (admin only).
func (h *UserManagementHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.Service.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.IsAdmin() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the admin"})
		return
	}

	if err := h.Service.Delete(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
