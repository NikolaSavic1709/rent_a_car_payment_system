package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
	"webshop/internal/database"
)

var jwtSecret = []byte("xDGnlTeEGXNeXnZDKn4yo17AL5f3bLLV8o4cz_avigGaUaMeoKeBIWNAnLVQ25609G2UmE-tD-8bZQpZfVbR6A")

type User struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Username string `db:"username"`
}

type AuthService struct {
	db database.Service // Assuming PostgresService interface is provided by your database package
}

func NewAuthService(db database.Service) *AuthService {
	return &AuthService{db: db}
}

func (a *AuthService) Authenticate(username, password string) (bool, error) {
	user, err := a.db.GetUserByUsernameAndPassword(username, password)

	if user == nil {
		return false, err
	}
	return true, nil
}

func (a *AuthService) GenerateToken(username, role string, userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"username": username,
		"role":     role,
		"userId":   userId,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}
		parts := strings.SplitN(tokenString, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing bearer token"})
			c.Abort()
			return
		}
		tokenString = parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")

			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		if claims["userId"] != nil {
			var userID int = int(claims["userId"].(float64))
			c.Set("userId", userID)
		}
		
		c.Set("username", claims["username"])
		c.Next()
	}
}

func (s *Server) ProtectedHandler(c *gin.Context) {
	username, _ := c.Get("username")
	fmt.Println("LOGIN3")
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the protected route!", "username": username})
}

func Authorize(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}
		parts := strings.SplitN(tokenString, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing bearer token"})
			c.Abort()
			return
		}
		tokenString = parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")

			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		fmt.Println(claims)
		if claims["role"] != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Set("role", claims["role"])
		c.Set("userId", claims["userId"])
		c.Set("username", claims["username"])
		c.Next()
	}
}
