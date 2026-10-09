package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSSHInfoDisabledAndEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/off", SSHInfo(nil))
	r.GET("/on", SSHInfo(func() any {
		return gin.H{"enabled": true, "host": "mud.example", "guest_enabled": false, "host_key_fingerprints": []string{"SHA256:abc"}}
	}))
	for _, tc := range []struct {
		path    string
		enabled bool
	}{
		{"/off", false},
		{"/on", true},
	} {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s status %d", tc.path, rec.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body["enabled"] != tc.enabled {
			t.Fatalf("%s body %s", tc.path, rec.Body.String())
		}
	}
}
