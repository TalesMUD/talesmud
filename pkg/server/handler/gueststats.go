package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/service"
)

// GuestStatsHandler exposes read-only guest analytics to admin users.
type GuestStatsHandler struct {
	StatsService service.GuestStatsService
	GuestService service.GuestService
}

// GetSummary returns aggregate guest statistics (total all-time, today, this week, active now, daily breakdown).
func (h *GuestStatsHandler) GetSummary(c *gin.Context) {
	summary, err := h.StatsService.GetSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load guest statistics"})
		return
	}
	// Override activeNow with a live count from the guest service (users table is authoritative)
	summary.ActiveNow = h.GuestService.CountActiveGuests()
	c.JSON(http.StatusOK, summary)
}

// GetDailyStats returns per-day session counts for the last N days (default 30, max 365).
func (h *GuestStatsHandler) GetDailyStats(c *gin.Context) {
	days := 30
	if d := c.Query("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n > 0 && n <= 365 {
			days = n
		}
	}
	stats, err := h.StatsService.GetDailyStats(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load daily statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
