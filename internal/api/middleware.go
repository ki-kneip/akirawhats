package api

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ctxUserID   = "userID"
	ctxUserRole = "userRole"
)

func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		secret := []byte(os.Getenv("JWT_SECRET"))
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set(ctxUserID, userID)
		if role, ok := claims["role"].(string); ok {
			c.Set(ctxUserRole, role)
		}
		c.Next()
	}
}

func issueToken(userID, role string) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour * 7).Unix(),
	})
	return token.SignedString(secret)
}

func adminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ctxUserRole)
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func getUserID(c *gin.Context) string {
	id, _ := c.Get(ctxUserID)
	s, _ := id.(string)
	return s
}

func getUserRole(c *gin.Context) string {
	role, _ := c.Get(ctxUserRole)
	s, _ := role.(string)
	return s
}
