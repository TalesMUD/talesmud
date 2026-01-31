package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"

	dbsqlite "github.com/talesmud/talesmud/pkg/db/sqlite"
	mud "github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/repository"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"
	"github.com/talesmud/talesmud/pkg/webui"
	"github.com/talesmud/talesmud/pkg/webuiplay"
)

// App ... main application structure
type App interface {
	Run()
}

type app struct {
	Router *gin.Engine
	Facade service.Facade
	mud    mud.MUDServer
}

// NewApp returns an application instance
// this is the primary stateless server providing an API interface
func NewApp() App {
	path := strings.TrimSpace(os.Getenv("SQLITE_PATH"))
	if path == "" {
		path = "talesmud.db"
	}
	client, err := dbsqlite.Open(path)
	if err != nil {
		log.WithError(err).Fatal("Failed to open SQLite database")
	}
	repos := repository.NewSQLiteFactory(client)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	scriptRunner := runner.NewMultiRunner()
	facade := service.NewFacade(repos, scriptRunner)
	mud := mud.New(facade)
	scriptRunner.SetServices(facade, mud.GameCtrl())

	return &app{
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

	npcSpawners := &handler.NPCSpawnersHandler{
		Service: app.Facade.NPCSpawnersService(),
	}

	dialogs := &handler.DialogsHandler{
		Service: app.Facade.DialogsService(),
	}

	charTemplates := &handler.CharacterTemplatesHandler{
		Repo:      app.Facade.CharacterTemplatesRepo(),
		ItemsRepo: app.Facade.ItemsService(),
	}

	lootTables := &handler.LootTablesHandler{
		Service: app.Facade.LootTablesService(),
	}

	backgroundsPath := strings.TrimSpace(os.Getenv("BACKGROUNDS_PATH"))
	if backgroundsPath == "" {
		backgroundsPath = "./uploads/backgrounds"
	}
	backgrounds := &handler.BackgroundsHandler{
		BasePath: backgroundsPath,
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

	serverSettings := &handler.ServerSettingsHandler{
		Service: app.Facade.ServerSettingsService(),
	}

	userMgmt := &handler.UserManagementHandler{
		Service: app.Facade.UsersService(),
	}

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "API is up and running")
	})

	// admin endpoints (basic auth for export/import)
	authorized := r.Group("/admin/", gin.BasicAuth(gin.Accounts{
		os.Getenv("ADMIN_USER"): os.Getenv("ADMIN_PASSWORD"),
	}))
	{
		authorized.GET("export", exp.Export)
		authorized.POST("import", exp.Import)
		authorized.GET("world", worldRenderer.Render)
	}

	// Protected API routes (JWT auth required)
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(app.Facade))
	{
		// Player-level routes (any authenticated user)

		// Characters
		protected.GET("characters", csh.GetCharacters)
		protected.GET("my-characters", csh.GetMyCharacters)
		protected.POST("characters", csh.PostCharacter)
		protected.GET("characters/:id", csh.GetCharacterByID)
		protected.DELETE("characters/:id", csh.DeleteCharacterByID)
		protected.PUT("characters/:id", csh.UpdateCharacterByID)
		protected.POST("newcharacter", csh.CreateNewCharacter)

		// Read-only game data (accessible to all authenticated users)
		protected.GET("rooms", rooms.GetRooms)
		protected.GET("rooms-vh", rooms.GetRoomValueHelp)
		protected.GET("rooms/:id", rooms.GetRoomByID)
		protected.GET("items", items.GetItems)
		protected.GET("items/:id", items.GetItemByID)
		protected.GET("scripts", scripts.GetScripts)
		protected.GET("script-types", scripts.GetScriptTypes)
		protected.GET("world/map", worldRenderer.Render)
		protected.GET("world/graph", worldRenderer.RenderGraphData)
		protected.GET("world/rooms-minimal", worldRenderer.GetMinimalRooms)
		protected.GET("npcs", npcs.GetNPCs)
		protected.GET("npcs/templates", npcs.GetNPCTemplates)
		protected.GET("npcs/:id", npcs.GetNPCByID)
		protected.GET("spawners", npcSpawners.GetSpawners)
		protected.GET("spawners/:id", npcSpawners.GetSpawnerByID)
		protected.GET("dialogs", dialogs.GetDialogs)
		protected.GET("dialogs/:id", dialogs.GetDialogByID)
		protected.GET("character-templates", charTemplates.GetCharacterTemplates)
		protected.GET("character-templates/:id", charTemplates.GetCharacterTemplateByID)
		protected.GET("character-templates/presets", charTemplates.GetCharacterTemplatePresets)
		protected.GET("loottables", lootTables.GetLootTables)
		protected.GET("loottables/:id", lootTables.GetLootTableByID)
		protected.GET("backgrounds", backgrounds.ListBackgrounds)
		protected.GET("settings", serverSettings.GetServerSettings)

		// User profile (any authenticated user can view/edit own profile)
		protected.GET("user", usr.GetUser)
		protected.PUT("user", usr.UpdateUser)

		// Creator-level routes (creator or admin role required)
		creator := protected.Group("")
		creator.Use(CreatorMiddleware())
		{
			// Rooms
			creator.POST("rooms", rooms.PostRoom)
			creator.PUT("rooms/:id", rooms.PutRoom)
			creator.DELETE("rooms/:id", rooms.DeleteRoom)

			// Items
			creator.POST("items", items.PostItem)
			creator.PUT("items/:id", items.UpdateItemByID)
			creator.DELETE("items/:id", items.DeleteItemByID)
			creator.POST("items/from-template/:templateId", items.CreateInstanceFromTemplate)

			// Scripts
			creator.POST("scripts", scripts.PostScript)
			creator.PUT("scripts/:id", scripts.PutScript)
			creator.DELETE("scripts/:id", scripts.DeleteScript)
			creator.POST("run-script/:id", scripts.ExecuteScript)

			// NPCs
			creator.POST("npcs", npcs.PostNPC)
			creator.PUT("npcs/:id", npcs.UpdateNPCByID)
			creator.DELETE("npcs/:id", npcs.DeleteNPCByID)
			creator.POST("npcs/:id/spawn", npcs.SpawnNPC)

			// NPC Spawners
			creator.POST("spawners", npcSpawners.PostSpawner)
			creator.PUT("spawners/:id", npcSpawners.UpdateSpawnerByID)
			creator.DELETE("spawners/:id", npcSpawners.DeleteSpawnerByID)

			// Dialogs
			creator.POST("dialogs", dialogs.PostDialog)
			creator.PUT("dialogs/:id", dialogs.UpdateDialogByID)
			creator.DELETE("dialogs/:id", dialogs.DeleteDialogByID)

			// Character Templates
			creator.POST("character-templates", charTemplates.PostCharacterTemplate)
			creator.PUT("character-templates/:id", charTemplates.UpdateCharacterTemplateByID)
			creator.DELETE("character-templates/:id", charTemplates.DeleteCharacterTemplateByID)
			creator.POST("character-templates/seed", charTemplates.SeedCharacterTemplates)

			// Loot Tables
			creator.POST("loottables", lootTables.PostLootTable)
			creator.PUT("loottables/:id", lootTables.UpdateLootTableByID)
			creator.DELETE("loottables/:id", lootTables.DeleteLootTableByID)
			creator.POST("loottables/:id/roll", lootTables.RollLootTable)

			// Backgrounds
			creator.POST("backgrounds/upload", backgrounds.UploadBackground)
			creator.DELETE("backgrounds/:filename", backgrounds.DeleteBackground)

			// Server Settings
			creator.PUT("settings", serverSettings.UpdateServerSettings)
		}

		// Admin-level routes (admin role required)
		adminAPI := protected.Group("admin/")
		adminAPI.Use(AdminMiddleware())
		{
			adminAPI.GET("users", userMgmt.GetAllUsers)
			adminAPI.PUT("users/:id/role", userMgmt.UpdateUserRole)
			adminAPI.POST("users/:id/ban", userMgmt.BanUser)
			adminAPI.POST("users/:id/unban", userMgmt.UnbanUser)
			adminAPI.DELETE("users/:id", userMgmt.DeleteUser)
		}
	}

	public := r.Group("/api/")
	{
		// Serve background images (public, no auth required)
		public.GET("backgrounds/:filename", backgrounds.ServeBackground)

		// Legacy endpoint for old character creation flow (returns hardcoded templates)
		public.GET("templates/characters", csh.GetCharacterTemplates)
		public.GET("item-slots", items.GetItemSlots)
		public.GET("item-qualities", items.GetItemQualities)
		public.GET("item-types", items.GetItemTypes)
		public.GET("item-subtypes", items.GetItemSubTypes)

		public.GET("room-of-the-day", rooms.GetRoomOfTheDay)

		// Public server info (no auth, used by MUD client)
		public.GET("server-info", serverSettings.GetServerInfo)
	}

	// Start MUD Server
	app.mud.Run()

	ws := r.Group("/ws")
	ws.Use(AuthMiddleware(app.Facade))
	ws.GET("", app.mud.HandleConnections)

	// Serve mud-client (game client) at /play
	r.Use(SPAMiddleware("/play", webuiplay.FS(), webuiplay.IndexFile))

	// Optional landing page from OS filesystem
	landingPath := strings.TrimSpace(os.Getenv("LANDING_PATH"))
	cwd, _ := os.Getwd()
	log.WithFields(log.Fields{
		"LANDING_PATH": landingPath,
		"cwd":          cwd,
	}).Info("Landing page configuration")
	r.Use(LandingMiddleware(landingPath))

	// Serve main app at /
	r.Use(SPAMiddleware("/", webui.FS(), webui.IndexFile))

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
