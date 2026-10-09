package mudserver

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/gamemode"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

// TestWebsocketReplaceClosesOldWith4001 pins the classic socket takeover:
// the newer socket receives the same welcome JSON, and the older socket
// is closed with 4001 "session replaced" without dropping the new one.
func TestWebsocketReplaceClosesOldWith4001(t *testing.T) {
	gin.SetMode(gin.TestMode)
	prev := gamemode.Current()
	t.Cleanup(func() {
		path := filepath.Join(t.TempDir(), "restore.yaml")
		body := "presentation: " + prev.Presentation + "\nauth: " + prev.Auth + "\n"
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Errorf("restore mode: %v", err)
			return
		}
		if err := gamemode.ApplyFile(path); err != nil {
			t.Errorf("restore mode: %v", err)
		}
	})
	mode := filepath.Join(t.TempDir(), "classic.yaml")
	if err := os.WriteFile(mode, []byte("presentation: classic\nauth: auth0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gamemode.ApplyFile(mode); err != nil {
		t.Fatal(err)
	}

	client, err := dbsqlite.Open(filepath.Join(t.TempDir(), "ws.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(client), nil)
	user := &entities.User{
		Entity:   &entities.Entity{ID: "user-ws"},
		RefID:    "ref-ws",
		Nickname: "Tester",
	}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatal(err)
	}

	mud := New(facade)
	r := gin.New()
	r.GET("/ws", func(c *gin.Context) {
		c.Set("user", user)
		mud.HandleConnections(c)
	})
	httpSrv := httptest.NewServer(r)
	t.Cleanup(httpSrv.Close)
	url := "ws" + strings.TrimPrefix(httpSrv.URL, "http") + "/ws"

	dial := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			t.Fatal(err)
		}
		return conn
	}
	want, err := json.Marshal(messages.NewRoomBasedMessage("", "Connected to [TalesMUD] ..."))
	if err != nil {
		t.Fatal(err)
	}
	want = append(want, '\n')

	readFrame := func(conn *websocket.Conn) ([]byte, error) {
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, payload, err := conn.ReadMessage()
		return payload, err
	}

	c1 := dial()
	t.Cleanup(func() { _ = c1.Close() })
	got1, err := readFrame(c1)
	if err != nil {
		t.Fatalf("first welcome: %v", err)
	}
	if string(got1) != string(want) {
		t.Fatalf("first welcome bytes\n got %q\nwant %q", got1, want)
	}

	c2 := dial()
	t.Cleanup(func() { _ = c2.Close() })

	var closed bool
	deadline := time.Now().Add(3 * time.Second)
	for !closed && time.Now().Before(deadline) {
		_ = c1.SetReadDeadline(time.Now().Add(time.Until(deadline)))
		_, _, err = c1.ReadMessage()
		if err == nil {
			continue
		}
		var closeErr *websocket.CloseError
		if !errors.As(err, &closeErr) {
			t.Fatalf("old socket error %T %v", err, err)
		}
		if closeErr.Code != closeSessionReplaced {
			t.Fatalf("close code %d reason %q", closeErr.Code, closeErr.Text)
		}
		if closeErr.Text != "session replaced" {
			t.Fatalf("close reason %q", closeErr.Text)
		}
		closed = true
	}
	if !closed {
		t.Fatal("old socket stayed open")
	}

	got2, err := readFrame(c2)
	if err != nil {
		t.Fatalf("second welcome: %v", err)
	}
	if string(got2) != string(want) {
		t.Fatalf("second welcome bytes\n got %q\nwant %q", got2, want)
	}

	_ = c2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, _, err = c2.ReadMessage()
	if err == nil {
		return
	}
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		t.Fatalf("new socket closed %d %q", closeErr.Code, closeErr.Text)
	}
}
