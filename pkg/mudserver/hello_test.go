package mudserver

import (
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/talesmud/talesmud/pkg/db/sqlite"
	"github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/mudserver/game/messages"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/service"
)

func TestWebsocketHelloNeverDispatchesCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "hello.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	facade := service.NewFacade(repository.NewSQLiteFactory(db), nil)
	user := &entities.User{Entity: &entities.Entity{ID: "user-hello"}, RefID: "ref-hello"}
	if _, err := facade.UsersService().Import(user); err != nil {
		t.Fatal(err)
	}
	mud := New(facade, nil).(*server)
	router := gin.New()
	router.GET("/ws", func(c *gin.Context) { c.Set("user", user); mud.HandleConnections(c) })
	httpServer := httptest.NewServer(router)
	t.Cleanup(httpServer.Close)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(httpServer.URL, "http")+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var welcome messages.MessageResponse
	if err := conn.ReadJSON(&welcome); err != nil {
		t.Fatal(err)
	}
	if welcome.Message != "Connected to [TalesMUD] ..." {
		t.Fatalf("welcome changed: %+v", welcome)
	}
	for _, text := range []string{"", "attack rat"} {
		if err := conn.WriteJSON(messages.IncomingMessage{Type: "hello", Proto: 1, Client: "test/1.0", Message: text}); err != nil {
			t.Fatal(err)
		}
		var hello messages.HelloMessage
		if err := conn.ReadJSON(&hello); err != nil {
			t.Fatal(err)
		}
		if hello.Type != "hello" || hello.Proto != 1 || hello.Status != "ok" || hello.Server == "" {
			t.Fatalf("hello: %+v", hello)
		}
		if len(mud.Game.OnMessageReceived()) != 0 {
			t.Fatal("hello was dispatched as a command")
		}
	}
	if err := conn.WriteJSON(messages.IncomingMessage{Message: "look"}); err != nil {
		t.Fatal(err)
	}
	select {
	case out := <-mud.Game.OnMessageReceived():
		msg, ok := out.(*messages.Message)
		if !ok || msg.Data != "look" {
			t.Fatalf("old input changed: %#v", out)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("classic command was not dispatched")
	}
}
