package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"

	"github.com/talesmud/talesmud/pkg/db"
	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	mud "github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/webui"
)

// App ... main application structure
type App interface {
	Run()
}

type app struct {

	// generic app base
	Router *gin.Engine
	db     *db.Client
	// owndnd specific
	Facade service.Facade
	mud    mud.MUDServer
}

// NewApp returns an application instance
// this is the primary stateless server providing an API interface
func NewApp() App {
	driver := strings.ToLower(strings.TrimSpace(os.Getenv("DB_DRIVER")))
	if driver == "" {
		if strings.TrimSpace(os.Getenv("SQLITE_PATH")) != "" {
			driver = "sqlite"
		} else {
			driver = "mongo"
		}
	}

	var repos repository.Factory
	switch driver {
	case "mongo":
		client := db.New(os.Getenv("MONGODB_DATABASE"))
		client.Connect(os.Getenv("MONGODB_CONNECTION_STRING"))
		repos = repository.NewMongoFactory(client)
	case "sqlite":
		path := strings.TrimSpace(os.Getenv("SQLITE_PATH"))
		if path == "" {
			path = "talesmud.db"
		}
		client, err := dbsqlite.Open(path)
		if err != nil {
			log.WithError(err).Fatal("Failed to open SQLite database")
		}
		repos = repository.NewSQLiteFactory(client)
	default:
		log.WithField("driver", driver).Fatal("Unsupported DB driver")
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	runner := runner.NewDefaultScriptRunner()
	facade := service.NewFacade(repos, runner)
	mud := mud.New(facade)
	runner.SetServices(facade, mud.GameCtrl())

	return &app{
		db:     nil,
		Router: r,
		Facade: facade,
		mud:    mud,
	}
}

// SetupRoutes ... Configures the routes
func (app *app) setupRoutes() {

	r := app.Router

	csh := &handler.CharactersHandler{
		app.Facade.CharactersService(),
	}

	usr := &handler.UsersHandler{
		app.Facade.UsersService(),
	}

	rooms := &handler.RoomsHandler{
		app.Facade.RoomsService(),
	}

	items := &handler.ItemsHandler{
		app.Facade.ItemsService(),
	}

	scripts := &handler.ScriptsHandler{
		app.Facade.ScriptsService(),
		app.Facade.Runner(),
	}

	npcs := &handler.NPCsHandler{
		Service: app.Facade.NPCsService(),
	}

	dialogs := &handler.DialogsHandler{
		Service: app.Facade.DialogsService(),
	}

	exp := &handler.ExportHandler{
		RoomsService:      app.Facade.RoomsService(),
		CharactersService: app.Facade.CharactersService(),
		UserService:       app.Facade.UsersService(),
		ItemsService:      app.Facade.ItemsService(),
		ScriptService:     app.Facade.ScriptsService(),
		NPCsService:       app.Facade.NPCsService(),
		DialogsService:    app.Facade.DialogsService(),
		PartiesService:    app.Facade.PartiesService(),
	}

	worldRenderer := &handler.WorldRendererHandler{
		RoomsService: app.Facade.RoomsService(),
	}

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "API is up and running")
	})

	// admin endpoints
	authorized := r.Group("/admin/", gin.BasicAuth(gin.Accounts{
		os.Getenv("ADMIN_USER"): os.Getenv("ADMIN_PASSWORD"),
	}))
	{
		authorized.GET("export", exp.Export)
		authorized.POST("import", exp.Import)
		authorized.GET("world", worldRenderer.Render)

	}

	// user,
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(app.Facade))
	{
		// CRUD
		protected.GET("characters", csh.GetCharacters)
		protected.POST("characters", csh.PostCharacter)
		protected.GET("characters/:id", csh.GetCharacterByID)
		protected.DELETE("characters/:id", csh.DeleteCharacterByID)
		protected.PUT("characters/:id", csh.UpdateCharacterByID)
		// special
		protected.POST("newcharacter", csh.CreateNewCharacter)

		protected.GET("rooms", rooms.GetRooms)
		protected.GET("rooms-vh", rooms.GetRoomValueHelp)

		protected.POST("rooms", rooms.PostRoom)
		protected.PUT("rooms/:id", rooms.PutRoom)
		protected.DELETE("rooms/:id", rooms.DeleteRoom)

		// items API should probably not be directly public
		protected.GET("items", items.GetItems)
		protected.POST("items", items.PostItem)
		protected.PUT("items/:id", items.UpdateItemByID)
		protected.DELETE("items/:id", items.DeleteItemByID)

		protected.GET("item-templates", items.GetItemTemplates)
		protected.POST("item-templates", items.PostItemTemplate)
		protected.PUT("item-templates/:id", items.UpdateItemTemplateByID)
		protected.DELETE("item-templates/:id", items.DeleteItemTemplateByID)

		protected.DELETE("item-create/:templateId", items.CreateItemFromTemplateID)

		// -- scripts
		protected.GET("scripts", scripts.GetScripts)
		protected.GET("script-types", scripts.GetScriptTypes)
		protected.POST("scripts", scripts.PostScript)
		protected.PUT("scripts/:id", scripts.PutScript)
		protected.DELETE("scripts/:id", scripts.DeleteScript)
		protected.POST("run-script/:id", scripts.ExecuteScript)

		protected.GET("world/map", worldRenderer.Render)

		protected.GET("user", usr.GetUser)
		protected.PUT("user", usr.UpdateUser)

		// NPCs
		protected.GET("npcs", npcs.GetNPCs)
		protected.POST("npcs", npcs.PostNPC)
		protected.GET("npcs/:id", npcs.GetNPCByID)
		protected.PUT("npcs/:id", npcs.UpdateNPCByID)
		protected.DELETE("npcs/:id", npcs.DeleteNPCByID)

		// Dialogs
		protected.GET("dialogs", dialogs.GetDialogs)
		protected.POST("dialogs", dialogs.PostDialog)
		protected.GET("dialogs/:id", dialogs.GetDialogByID)
		protected.PUT("dialogs/:id", dialogs.UpdateDialogByID)
		protected.DELETE("dialogs/:id", dialogs.DeleteDialogByID)
	}

	public := r.Group("/api/")
	{
		public.GET("templates/characters", csh.GetCharacterTemplates)
		public.GET("item-slots", items.GetItemSlots)
		public.GET("item-qualities", items.GetItemQualities)
		public.GET("item-types", items.GetItemTypes)
		public.GET("item-subtypes", items.GetItemSubTypes)

		public.GET("room-of-the-day", rooms.GetRoomOfTheDay)

	}

	// Start MUD Server
	app.mud.Run()

	ws := r.Group("/ws")
	ws.Use(AuthMiddleware(app.Facade))
	ws.GET("", app.mud.HandleConnections)

	//staticHandler := static.ServeRoot("/app/*filepath", "public/app/public/")

	//staticHandler := gin.WrapH(http.Handler(http.FileServer(http.Dir("public/app/public"))))
	//r.GET("/app/*any", staticHandler)
	//r.NoRoute(staticHandler)
	r.Use(SPAMiddleware("/", webui.FS(), webui.IndexFile))

	//r.Use(staticHandler)

}

// Run ... starts the server
func (app *app) Run() {

	app.setupRoutes()

	// read port from env file
	port := os.Getenv("PORT")

	server := fmt.Sprintf("0.0.0.0:%v", port)

	// setup CORS handler
	corsHandler := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "DELETE", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))(app.Router)

	log.WithField("PORT", port).Info(fmt.Sprintf("TalesMUD Server is running, listening on port %v", port))
	log.Fatal(http.ListenAndServe(server, corsHandler))
}
