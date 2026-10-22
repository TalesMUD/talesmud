package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/talesmud/talesmud/pkg/entities/settings"
	"github.com/talesmud/talesmud/pkg/ruleset"
)

type stubServerSettings struct {
	name string
}

func (s stubServerSettings) Get() (*settings.ServerSettings, error) {
	return &settings.ServerSettings{ServerName: s.name}, nil
}

func (stubServerSettings) Update(*settings.ServerSettings) error { return nil }

func TestServerInfoDefaultTheme(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ruleset.Reset()
	t.Cleanup(ruleset.Reset)

	r := gin.New()
	h := &ServerSettingsHandler{Service: stubServerSettings{name: "Layout"}}
	r.GET("/api/server-info", h.GetServerInfo)

	get := func() map[string]any {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/server-info", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
		}
		var body map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		return body
	}

	body := get()
	if body["serverName"] != "Layout" || body["defaultTheme"] != ruleset.DefaultThemeDarkFantasy {
		t.Fatalf("body %#v", body)
	}
	if err := ruleset.LoadBytes([]byte("ui:\n  default_theme: gilded\n")); err != nil {
		t.Fatal(err)
	}
	body = get()
	if body["defaultTheme"] != "gilded" {
		t.Fatalf("body %#v", body)
	}
}
