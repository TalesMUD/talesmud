package server

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	e "github.com/talesmud/talesmud/pkg/entities"
	"github.com/talesmud/talesmud/pkg/service"
)

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

// getPemCert is a function used to get the PEM certificate from a JWT token. It retrieves the certificate
// by calling the Auth0 JWKS endpoint and parsing the JSON response to find the certificate that matches the key ID
// specified in the JWT token header.
func getPemCert(token *jwt.Token) (string, error) {

	// Initialize an empty string to hold the certificate
	cert := ""

	// Make a GET request to the JWKS endpoint specified in the environment variables
	resp, err := http.Get(os.Getenv("AUTH0_WK_JWKS"))

	// If the request results in an error, return the empty certificate string and the error
	if err != nil {
		return cert, err
	}

	// Make sure to close the response body when the function returns
	defer resp.Body.Close()

	// Initialize a jwks structure to hold the JSON response
	var jwks = jwks{}

	// Decode the JSON response into the jwks structure
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	// If decoding the JSON results in an error, return the empty certificate string and the error
	if err != nil {
		return cert, err
	}

	// Iterate over the keys in the JWKS response
	for k := range jwks.Keys {
		// If the key ID of the current key matches the key ID specified in the JWT token header,
		// construct the certificate string using the certificate value (x5c) from the JWKS key
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	// If no matching key was found in the JWKS, return an error
	if cert == "" {
		err := errors.New("Unable to find appropriate key")
		return cert, err
	}

	// If a matching key was found, return the certificate string and nil error
	return cert, nil
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
// If the token is valid, it sets the user ID and user in the gin context.
func AuthMiddleware(facade service.Facade) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info("GIN JWT MIDDLEWARE")

		keyFunc := getKeyFunc()

		var token *jwt.Token
		var err error

		if fromQuery, ok := c.GetQuery("access_token"); ok {
			log.Info("Found access token in query param")
			token, err = jwt.Parse(fromQuery, keyFunc)
		} else {
			log.Info("Found access token in http header")
			token, err = request.ParseFromRequest(c.Request, request.AuthorizationHeaderExtractor, keyFunc)
		}

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
