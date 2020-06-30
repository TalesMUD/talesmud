package server

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-contrib/static"
	log "github.com/sirupsen/logrus"

	"net/http"
	"strings"

	"encoding/base64"

	"github.com/buger/jsonparser"

	"github.com/talesmud/talesmud/pkg/db"
	mud "github.com/talesmud/talesmud/pkg/mudserver"
	"github.com/talesmud/talesmud/pkg/scripts/runner"
	"github.com/talesmud/talesmud/pkg/server/handler"
	"github.com/talesmud/talesmud/pkg/service"

	"errors"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/handlers"
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
	facade service.Facade
	mud    mud.MUDServer
}

// NewApp returns an application instance
// this is the primary stateless server providing an API interface
func NewApp() App {

	db := db.New(os.Getenv("MONGODB_DATABASE"))
	db.Connect(os.Getenv("MONGODB_CONNECTION_STRING"))

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	runner := runner.NewDefaultScriptRunner()
	facade := service.NewFacade(db, runner)
	mud := mud.New(facade)
	runner.SetServices(facade, mud.GameCtrl())

	return &app{
		db:     db,
		Router: r,
		facade: facade,
		mud:    mud,
	}
}

/// AUTH0 handling
type jwks struct {
	Keys []webKeys `json:"keys"`
}
type webKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get(os.Getenv("AUTH0_WK_JWKS"))

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

func (app *app) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info("GIN JWT MIDDLEWARE")
		r := c.Request

		keyFunc := func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			aud := os.Getenv("AUTH0_AUDIENCE")
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New("Invalid audience")
			}
			// Verify 'iss' claim
			iss := os.Getenv("AUTH0_DOMAIN")
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer")
			}

			cert, err := getPemCert(token)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		}

		var token *jwt.Token
		var err error

		if fromQuery, ok := c.GetQuery("access_token"); ok {
			log.Info("Found access token in query param")
			token, err = jwt.Parse(fromQuery, keyFunc)

		} else {
			log.Info("Found access token in http header")
			token, err = request.ParseFromRequest(r, request.AuthorizationHeaderExtractor, keyFunc)
		}

		if err != nil {
			fmt.Println(err)
			fmt.Println("Token is not valid:", token)

			c.AbortWithStatus(401)

		} else {

			// set userid if not already in context
			if _, ok := c.Get("userid"); !ok {
				splitted := strings.Split(token.Raw, ".")
				if decoded, err := base64.RawStdEncoding.DecodeString(splitted[1]); err == nil {
					if sub, err := jsonparser.GetString(decoded, "sub"); err == nil {
						c.Set("userid", sub)
					} else {
						log.WithError(err).Error("Could not get sub part from JSON")
					}
				} else {
					//TODO: remove token logging
					log.WithError(err).WithField("RawToken", token.Raw).Error("Could not decode token part")
				}
			}

			if _, ok := c.Get("user"); !ok {
				if id, exists := c.Get("userid"); exists {
					if user, err := app.facade.UsersService().FindOrCreateNewUser(id.(string)); err == nil {
						log.WithField("UserID", user.ID).Debug("Set user in Context")
						c.Set("user", user)
					}
				}
			}

			c.Set("token", token)
			c.Next()
		}
	}
}

// SetupRoutes ... Configures the routes
func (app *app) setupRoutes() {

	r := app.Router

	csh := &handler.CharactersHandler{
		app.facade.CharactersService(),
	}

	usr := &handler.UsersHandler{
		app.facade.UsersService(),
	}

	rooms := &handler.RoomsHandler{
		app.facade.RoomsService(),
	}

	items := &handler.ItemsHandler{
		app.facade.ItemsService(),
	}

	scripts := &handler.ScriptsHandler{
		app.facade.ScriptsService(),
		app.facade.Runner(),
	}

	exp := &handler.ExportHandler{
		RoomsService:      app.facade.RoomsService(),
		CharactersService: app.facade.CharactersService(),
		UserService:       app.facade.UsersService(),
		ItemsService:      app.facade.ItemsService(),
		ScriptService:     app.facade.ScriptsService(),
	}

	worldRenderer := &handler.WorldRendererHandler{
		RoomsService: app.facade.RoomsService(),
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
	protected.Use(app.authMiddleware())
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
	}

	public := r.Group("/api/")
	{
		public.GET("templates/characters", csh.GetCharacterTemplates)
		public.GET("item-slots", items.GetItemSlots)
		public.GET("item-qualities", items.GetItemQualities)
		public.GET("item-types", items.GetItemTypes)
		public.GET("item-subtypes", items.GetItemSubTypes)
	}

	// Start MUD Server
	app.mud.Run()

	ws := r.Group("/ws")
	ws.Use(app.authMiddleware())
	ws.GET("", app.mud.HandleConnections)

	//staticHandler := static.ServeRoot("/app/*filepath", "public/app/public/")

	//staticHandler := gin.WrapH(http.Handler(http.FileServer(http.Dir("public/app/public"))))
	//r.GET("/app/*any", staticHandler)
	//r.NoRoute(staticHandler)
	r.Use(middleware("/", "./public/app/public"))

	//r.Use(staticHandler)

}

func middleware(urlPrefix, spaDirectory string) gin.HandlerFunc {
	directory := static.LocalFile(spaDirectory, true)
	fileserver := http.FileServer(directory)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if directory.Exists(urlPrefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		} else {
			c.Request.URL.Path = "/"
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}

// Run ... starts the server
func (app *app) Run() {

	app.setupRoutes()

	port := 8010
	server := fmt.Sprintf("0.0.0.0:%v", port)

	corsHandler := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "DELETE", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))(app.Router)

	log.WithField("PORT", port).Info(fmt.Sprintf("TalesMUD Server is running, listening on port %v", port))
	log.Fatal(http.ListenAndServe(server, corsHandler))
}
