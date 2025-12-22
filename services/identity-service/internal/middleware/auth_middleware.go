package middleware

import (
	"net/http"
	"strings"

	"github.com/Caesarsage/bankflow/identity-service/internal/models"
	"github.com/Caesarsage/bankflow/identity-service/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtManager *jwt.JWTManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing_token",
				Message: "Authorization token required",
			})
			ctx.Abort()
			return
		}

		// Bearer format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid_token_format",
				Message: "Authorization header must be Bearer {token}",
			})
			ctx.Abort()
			return
		}

		// Validate token
		token := parts[1]
		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid_token",
				Message: err.Error(),
			})
			ctx.Abort()
			return
		}

		// Set user info in context
		ctx.Set("user_id", claims.UserID)
		ctx.Set("email", claims.Email)

		ctx.Next()
	}
}

// CORS middleware
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
