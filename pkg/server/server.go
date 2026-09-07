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
	"github.com/talesmud/talesmud/pkg/service/groq"
	"github.com/talesmud/talesmud/pkg/webui"
	"github.com/talesmud/talesmud/pkg/webuiplay"
)

// App ... main application structure
type App interface {
	Run()
}

type app struct {
	Router  *gin.Engine
	Facade  service.Facade
	mud     mud.MUDServer
	tickets *TicketStore
}

func trustedProxies() []string {
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))
	if raw == "" {
		return []string{"127.0.0.1", "::1"}
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"127.0.0.1", "::1"}
	}
	return out
}

func adminAuthMiddleware() gin.HandlerFunc {
	user := strings.TrimSpace(os.Getenv("ADMIN_USER"))
	pass := strings.TrimSpace(os.Getenv("ADMIN_PASSWORD"))
	if user == "" || pass == "" {
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "admin import/export credentials are not configured",
			})
		}
	}
	if strings.EqualFold(os.Getenv("GIN_MODE"), "release") && user == "admin" && pass == "admin" {
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "admin import/export credentials are insecure",
			})
		}
	}
	return gin.BasicAuth(gin.Accounts{user: pass})
}

func allowedCORSOrigins() []string {
	origins := []string{
		"https://veilspan.com",
		"https://www.veilspan.com",
		"https://talesmud.io",
		"https://www.talesmud.io",
		"http://localhost:5000",
		"http://127.0.0.1:5000",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
		"http://localhost:8010",
		"http://127.0.0.1:8010",
	}

	for _, origin := range strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
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
	if err := r.SetTrustedProxies(trustedProxies()); err != nil {
		log.WithError(err).Warn("Failed to set trusted proxies; using Gin defaults")
	}
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		path := param.Path
		if i := strings.Index(path, "?"); i >= 0 {
			path = path[:i]
		}
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			path,
		)
	}))
	r.Use(gin.Recovery())

	scriptRunner := runner.NewMultiRunner()
	facade := service.NewFacade(repos, scriptRunner)
	mudSrv := mud.New(facade, allowedCORSOrigins())
	scriptRunner.SetServices(facade, mudSrv.GameCtrl())

	return &app{
		Router:  r,
		Facade:  facade,
		mud:     mudSrv,
		tickets: NewTicketStore(),
	}
}

// SetupRoutes ... Configures the routes
func (app *app) setupRoutes() {

	r := app.Router

	csh := &handler.CharactersHandler{
		Service: app.Facade.CharactersService(),
	}
	characterMap := &handler.CharacterMapHandler{
		Characters: app.Facade.CharactersService(),
		Rooms:      app.Facade.RoomsService(),
	}

	usr := &handler.UsersHandler{
		Service: app.Facade.UsersService(),
	}

	rooms := &handler.RoomsHandler{
		Service: app.Facade.RoomsService(),
		Facade:  app.Facade,
	}

	items := &handler.ItemsHandler{
		Service: app.Facade.ItemsService(),
		Facade:  app.Facade,
	}

	scripts := &handler.ScriptsHandler{
		Service: app.Facade.ScriptsService(),
		Runner:  app.Facade.Runner(),
		Facade:  app.Facade,
	}

	npcs := &handler.NPCsHandler{
		Service: app.Facade.NPCsService(),
		Facade:  app.Facade,
	}

	npcSpawners := &handler.NPCSpawnersHandler{
		Service: app.Facade.NPCSpawnersService(),
		Facade:  app.Facade,
	}

	dialogs := &handler.DialogsHandler{
		Service: app.Facade.DialogsService(),
		Facade:  app.Facade,
	}

	charTemplates := &handler.CharacterTemplatesHandler{
		Repo:      app.Facade.CharacterTemplatesRepo(),
		ItemsRepo: app.Facade.ItemsService(),
	}

	lootTables := &handler.LootTablesHandler{
		Service: app.Facade.LootTablesService(),
		Facade:  app.Facade,
	}

	questsHandler := &handler.QuestsHandler{
		Service:           app.Facade.QuestsService(),
		CharactersService: app.Facade.CharactersService(),
		Facade:            app.Facade,
	}

	skillsHandler := &handler.SkillsHandler{
		Service: app.Facade.SkillsService(),
	}

	generate := &handler.GenerateHandler{
		GroqClient: groq.NewClient(),
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
	worldValidation := &handler.WorldValidationHandler{
		Service: service.NewWorldValidationService(app.Facade),
	}

	serverSettings := &handler.ServerSettingsHandler{
		Service: app.Facade.ServerSettingsService(),
	}

	userMgmt := &handler.UserManagementHandler{
		Service: app.Facade.UsersService(),
	}

	guestStats := &handler.GuestStatsHandler{
		StatsService: app.Facade.GuestStatsService(),
		GuestService: app.Facade.GuestService(),
	}

	validationHandler := &handler.ValidationHandler{
		Facade: app.Facade,
	}

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "API is up and running")
	})

	// admin endpoints (basic auth for export/import)
	authorized := r.Group("/admin/", adminAuthMiddleware())
	{
		authorized.GET("export", exp.Export)
		authorized.POST("import", exp.Import)
	}

	// Protected API routes (JWT auth required)
	protected := r.Group("/api/")
	protected.Use(AuthMiddleware(app.Facade))
	{
		// Player-level routes (any authenticated user)

		// Characters
		protected.GET("characters", csh.GetCharacters)
		protected.GET("my-characters", csh.GetMyCharacters)
		protected.GET("characters/:id", csh.GetCharacterByID)
		protected.GET("characters/:id/map", characterMap.GetCharacterMap)
		protected.DELETE("characters/:id", csh.DeleteCharacterByID)
		protected.PUT("characters/:id", csh.UpdateCharacterByID)
		protected.POST("newcharacter", csh.CreateNewCharacter)
		protected.POST("ws-ticket", app.tickets.IssueWSTicket)

		// AI-powered generation
		protected.POST("generate/character", generate.GenerateCharacter)

		protected.GET("character-templates", charTemplates.GetCharacterTemplates)
		protected.GET("character-templates/:id", charTemplates.GetCharacterTemplateByID)
		protected.GET("character-templates/presets", charTemplates.GetCharacterTemplatePresets)
		protected.GET("quest-progress/:characterId", questsHandler.GetQuestLog)
		protected.POST("quest-progress/:characterId/accept/:questId", questsHandler.AcceptQuest)
		protected.POST("quest-progress/:characterId/abandon/:questId", questsHandler.AbandonQuest)
		protected.POST("quest-progress/:characterId/complete/:questId", questsHandler.CompleteQuest)

		// User profile (any authenticated user can view/edit own profile)
		protected.GET("user", usr.GetUser)
		protected.PUT("user", usr.UpdateUser)

		// Creator-level routes (creator or admin role required)
		creator := protected.Group("")
		creator.Use(CreatorMiddleware())
		{
			creator.POST("characters", csh.PostCharacter)

			creator.GET("rooms", rooms.GetRooms)
			creator.GET("rooms-vh", rooms.GetRoomValueHelp)
			creator.GET("rooms/:id", rooms.GetRoomByID)
			creator.GET("items", items.GetItems)
			creator.GET("items/:id", items.GetItemByID)
			creator.GET("scripts", scripts.GetScripts)
			creator.GET("script-types", scripts.GetScriptTypes)
			creator.GET("world/graph", worldRenderer.RenderGraphData)
			creator.GET("world/rooms-minimal", worldRenderer.GetMinimalRooms)
			creator.GET("npcs", npcs.GetNPCs)
			creator.GET("npcs/templates", npcs.GetNPCTemplates)
			creator.GET("npcs/:id", npcs.GetNPCByID)
			creator.GET("spawners", npcSpawners.GetSpawners)
			creator.GET("spawners/:id", npcSpawners.GetSpawnerByID)
			creator.GET("dialogs", dialogs.GetDialogs)
			creator.GET("dialogs/:id", dialogs.GetDialogByID)
			creator.GET("loottables", lootTables.GetLootTables)
			creator.GET("loottables/:id", lootTables.GetLootTableByID)
			creator.GET("backgrounds", backgrounds.ListBackgrounds)
			creator.GET("settings", serverSettings.GetServerSettings)
			creator.GET("quests", questsHandler.GetQuests)
			creator.GET("quests/:id", questsHandler.GetQuestByID)
			creator.GET("skills", skillsHandler.GetSkills)
			creator.GET("skills/:id", skillsHandler.GetSkillByID)

			// Rooms
			creator.POST("rooms", rooms.PostRoom)
			creator.PUT("rooms/:id", rooms.PutRoom)
			creator.DELETE("rooms/:id", rooms.DeleteRoom)
			creator.PUT("world/rooms-coords", worldRenderer.BatchUpdateCoords)
			creator.GET("world/validation", worldValidation.GetWorldValidation)

			// Creator quality diagnostics
			creator.GET("diagnostics/world", validationHandler.WorldDiagnostics)
			creator.POST("validate/:entityType", validationHandler.ValidateEntity)
			creator.POST("preview/dialog", validationHandler.PreviewDialog)
			creator.POST("preview/quest", validationHandler.PreviewQuest)
			creator.POST("preview/room", validationHandler.PreviewRoom)
			creator.POST("preview/merchant", validationHandler.PreviewMerchant)

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

			// Quests
			creator.POST("quests", questsHandler.PostQuest)
			creator.PUT("quests/:id", questsHandler.UpdateQuestByID)
			creator.DELETE("quests/:id", questsHandler.DeleteQuestByID)

			// Skills
			creator.POST("skills", skillsHandler.PostSkill)
			creator.PUT("skills/:id", skillsHandler.UpdateSkillByID)
			creator.DELETE("skills/:id", skillsHandler.DeleteSkillByID)

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

			// Guest session analytics
			adminAPI.GET("stats/guests", guestStats.GetSummary)
			adminAPI.GET("stats/guests/daily", guestStats.GetDailyStats)
		}
	}

	public := r.Group("/api/")
	{
		// Serve background images (public, no auth required)
		public.GET("backgrounds/:filename", backgrounds.ServeBackground)

		portraitsPath := os.Getenv("PORTRAITS_PATH")
		if portraitsPath == "" {
			portraitsPath = "./uploads/portraits"
		}
		portraitsHandler := &handler.BackgroundsHandler{BasePath: portraitsPath}
		public.GET("portraits/:filename", portraitsHandler.ServeBackground)

		itemsPath := os.Getenv("ITEM_ART_PATH")
		if itemsPath == "" {
			itemsPath = "./uploads/items"
		}
		itemsHandler := &handler.BackgroundsHandler{BasePath: itemsPath}
		public.GET("item-art/:filename", itemsHandler.ServeBackground)
		public.HEAD("item-art/:filename", itemsHandler.ServeBackground)

		// Legacy endpoint for old character creation flow (returns hardcoded templates)
		public.GET("templates/characters", csh.GetCharacterTemplates)
		public.GET("item-slots", items.GetItemSlots)
		public.GET("item-qualities", items.GetItemQualities)
		public.GET("item-types", items.GetItemTypes)
		public.GET("item-subtypes", items.GetItemSubTypes)

		public.GET("room-of-the-day", rooms.GetRoomOfTheDay)

		// Public server info (no auth, used by MUD client)
		public.GET("server-info", serverSettings.GetServerInfo)

		// Guest session creation (public, no auth required)
		guest := &handler.GuestHandler{
			GuestService: app.Facade.GuestService(),
		}
		public.POST("guest", guest.CreateGuestSession)
	}

	// Start MUD Server
	app.mud.Run()

	// Start guest cleanup loop (removes expired guest accounts every 5 minutes)
	app.Facade.GuestService().StartCleanupLoop()

	ws := r.Group("/ws")
	ws.Use(WSAuthMiddleware(app.Facade, app.tickets))
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
		handlers.AllowedOrigins(allowedCORSOrigins()))(app.Router)

	log.WithField("PORT", port).Info(fmt.Sprintf("TalesMUD Server is running, listening on port %v", port))
	log.Fatal(http.ListenAndServe(server, corsHandler))
}
