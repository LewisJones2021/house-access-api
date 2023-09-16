package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lewisjones2021/house-access-api/helpers"
)

// auth validates token and authorizes users.
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {

		// extract the token from the request header.
		clientToken := c.Request.Header.Get("token")

		// check if no authorization header is provided.
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("no authorisation header provided")})
			// abort the request.
			c.Abort()
			return
		}

		// validate the client token using the ValidateToken function from the helpers package.
		claims, err := helpers.ValidateToken(clientToken)

		// if there's an error during token validation, return an error response.
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		// set user claims (email, name, uid) as Gin context values for use in subsequent handlers.
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Set("uid", claims.Uid)

		// continue processing the request to the next handler.
		c.Next()
	}
}
