package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/service"
)

// jwksCache holds cached JWKS keys to avoid fetching from Auth0 on every request.
var jwksCache struct {
	sync.RWMutex
	keys      []webKeys
	fetchedAt time.Time
}

const jwksCacheTTL = 1 * time.Hour

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

// fetchJWKS retrieves JWKS keys from Auth0, using an in-memory cache with TTL.
func fetchJWKS() ([]webKeys, error) {
	jwksCache.RLock()
	if len(jwksCache.keys) > 0 && time.Since(jwksCache.fetchedAt) < jwksCacheTTL {
		keys := jwksCache.keys
		jwksCache.RUnlock()
		return keys, nil
	}
	jwksCache.RUnlock()

	jwksCache.Lock()
	defer jwksCache.Unlock()

	// Double-check after acquiring write lock
	if len(jwksCache.keys) > 0 && time.Since(jwksCache.fetchedAt) < jwksCacheTTL {
		return jwksCache.keys, nil
	}

	log.Info("Fetching JWKS keys from Auth0")
	resp, err := http.Get(os.Getenv("AUTH0_WK_JWKS"))
	if err != nil {
		// Return stale cache if available
		if len(jwksCache.keys) > 0 {
			log.WithError(err).Warn("Failed to refresh JWKS, using stale cache")
			return jwksCache.keys, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var parsed jwks
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		if len(jwksCache.keys) > 0 {
			log.WithError(err).Warn("Failed to decode JWKS, using stale cache")
			return jwksCache.keys, nil
		}
		return nil, err
	}

	jwksCache.keys = parsed.Keys
	jwksCache.fetchedAt = time.Now()
	log.WithField("keyCount", len(parsed.Keys)).Info("JWKS keys cached")
	return parsed.Keys, nil
}

// getPemCert retrieves the PEM certificate matching the JWT token's key ID from cached JWKS keys.
func getPemCert(token *jwt.Token) (string, error) {
	keys, err := fetchJWKS()
	if err != nil {
		return "", err
	}

	for _, key := range keys {
		if token.Header["kid"] == key.Kid {
			return "-----BEGIN CERTIFICATE-----\n" + key.X5c[0] + "\n-----END CERTIFICATE-----", nil
		}
	}

	return "", errors.New("Unable to find appropriate key")
}

// getKeyFunc returns a function to be used as the jwt.Keyfunc for JWT token validation.
// It verifies the 'aud' and 'iss' claims and extracts the PEM certificate.
func getKeyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
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
}

// handleTokenError handles the case where the JWT token is invalid.
// It logs the error and aborts the gin context with a 401 status.
func handleTokenError(c *gin.Context, err error, token *jwt.Token) {
	fmt.Println(err)
	fmt.Println("Token is not valid:", token)

	c.AbortWithStatus(401)
}

// handleTokenSuccess handles the case where the JWT token is valid.
// It sets the user ID and user in the gin context, if they don't already exist.
// It also enforces ban status — banned users are rejected with 403.
func handleTokenSuccess(c *gin.Context, token *jwt.Token, facade service.Facade) {
	// set userid if not already in context
	if _, ok := c.Get("userid"); !ok {
		setUserId(c, token)
	}

	if _, ok := c.Get("user"); !ok {
		setUser(c, facade)
	}

	// Ban enforcement: reject banned users at the auth layer
	if usr, exists := c.Get("user"); exists {
		if user, ok := usr.(*e.User); ok && user.IsBanned {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Your account has been banned",
			})
			return
		}
	}

	c.Set("token", token)
	c.Next()
}

// setUserId sets the user ID in the gin context.
// It decodes the JWT token and extracts the 'sub' claim.
func setUserId(c *gin.Context, token *jwt.Token) {
	splitted := strings.Split(token.Raw, ".")
	// JWTs use URL-safe base64 encoding (RawURLEncoding), not standard base64
	if decoded, err := base64.RawURLEncoding.DecodeString(splitted[1]); err == nil {
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

// setUser sets the user in the gin context.
// It retrieves the user ID from the context and calls the facade's UsersService to find or create a new user.
func setUser(c *gin.Context, facade service.Facade) {
	if id, exists := c.Get("userid"); exists {
		if user, err := facade.UsersService().FindOrCreateNewUser(id.(string)); err == nil {
			log.WithField("UserID", user.ID).Debug("Set user in Context")
			c.Set("user", user)
		}
	}
}

// AuthMiddleware is a gin middleware function for authentication.
// It verifies the JWT token from the query parameter or the authorization header.
// Supports both Auth0 JWTs and guest HMAC tokens (tried first for fast validation).
func AuthMiddleware(facade service.Facade) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info("GIN JWT MIDDLEWARE")

		// Extract token string from query param or Authorization header
		var tokenStr string
		if fromQuery, ok := c.GetQuery("access_token"); ok {
			tokenStr = fromQuery
		} else {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenStr == "" {
			c.AbortWithStatus(401)
			return
		}

		// Try guest token validation first (fast HMAC check)
		if guestSvc := facade.GuestService(); guestSvc != nil {
			if userID, err := guestSvc.ValidateGuestToken(tokenStr); err == nil {
				if user, err := facade.UsersService().FindByID(userID); err == nil {
					// Check if guest session has expired
					if user.IsGuest && !user.GuestExpiresAt.IsZero() && time.Now().After(user.GuestExpiresAt) {
						c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
							"error": "Guest session expired",
						})
						return
					}
					c.Set("userid", user.RefID)
					c.Set("user", user)
					c.Next()
					return
				}
			}
		}

		// Fall back to Auth0 JWT validation
		keyFunc := getKeyFunc()

		var token *jwt.Token
		var err error

		token, err = jwt.Parse(tokenStr, keyFunc)

		if err != nil {
			handleTokenError(c, err, token)
		} else {
			handleTokenSuccess(c, token, facade)
		}
	}
}

// CreatorMiddleware requires the authenticated user to have creator or admin role.
func CreatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if usr, exists := c.Get("user"); exists {
			if user, ok := usr.(*e.User); ok && user.IsCreator() {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Creator access required",
		})
	}
}

// AdminMiddleware requires the authenticated user to have admin role.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if usr, exists := c.Get("user"); exists {
			if user, ok := usr.(*e.User); ok && user.IsAdmin() {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Admin access required",
		})
	}
}
