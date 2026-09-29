package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CrossOrigin(allowedOrigins []string, exposedHeaders ...string) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", ClientIdentificationHeader, "Idempotency-Key", RequestIDHeader},
		ExposeHeaders:    append([]string{RequestIDHeader}, exposedHeaders...),
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
