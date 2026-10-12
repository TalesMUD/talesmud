package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/devicecode"
)

// DeviceAPI exposes app starts and polls on the same store as SSH activation.
type DeviceAPI struct {
	Devices     *devicecode.Store
	ActivateURL string
	Issue       func(string) (string, int, error)
}

func deviceHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}

func (a *DeviceAPI) Start(c *gin.Context) {
	deviceHeaders(c)
	var body struct {
		Client        string `json:"client"`
		ClientVersion string `json:"client_version"`
		Proto         int    `json:"proto"`
	}
	if !bindSSH(c, &body) {
		return
	}
	if strings.TrimSpace(body.Client) == "" || len(body.Client) > 64 || len(body.ClientVersion) > 64 || body.Proto != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	verify, err := url.Parse(a.ActivateURL)
	if err != nil || verify.Host == "" || (verify.Scheme != "https" && verify.Scheme != "http") || verify.User != nil || a.Devices == nil || a.Issue == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	code, err := a.Devices.BeginApp(c.ClientIP(), body.Client, body.ClientVersion)
	if errors.Is(err, devicecode.ErrLimited) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many attempts"})
		return
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
		return
	}
	query := verify.Query()
	query.Set("code", code.UserCode)
	verify.RawQuery = query.Encode()
	c.JSON(http.StatusOK, gin.H{
		"user_code": code.UserCode, "device_code": code.DeviceCode,
		"verification_uri": a.ActivateURL, "verification_uri_complete": verify.String(),
		"expires_in": code.ExpiresIn, "interval": code.Interval,
	})
}

func (a *DeviceAPI) Poll(c *gin.Context) {
	deviceHeaders(c)
	var body struct {
		DeviceCode string `json:"device_code"`
	}
	if !bindSSH(c, &body) {
		return
	}
	token, expires, err := a.Devices.Poll(body.DeviceCode, c.ClientIP(), a.Issue)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"token": token, "expires_in": expires})
	case errors.Is(err, devicecode.ErrPending), errors.Is(err, devicecode.ErrSlowDown), errors.Is(err, devicecode.ErrDenied), errors.Is(err, devicecode.ErrExpired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "unavailable"})
	}
}
