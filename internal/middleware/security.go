package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Définition centralisée de la politique CSP
var (
	trustedStyles = []string{
		"'self'",
		"'unsafe-inline'",
		"https://cdn.jsdelivr.net/npm/@picocss/",
	}
	trustedScripts = []string{
		"'self'",
		"'unsafe-inline'",
		"'unsafe-eval'",
		"https://unpkg.com/htmx.org@1.9.10",
		"https://unpkg.com/htmx.org@1.9.10/",
	}
)

// SecurityHeaders injecte les en-têtes de sécurité HTTP de base recommandés
func SecurityHeaders() gin.HandlerFunc {
	// Construction de la directive CSP une seule fois au démarrage
	cspPolicy := fmt.Sprintf(
		"default-src 'self'; style-src %s; script-src %s; img-src 'self' data:;",
		strings.Join(trustedStyles, " "),
		strings.Join(trustedScripts, " "),
	)

	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", cspPolicy)
		c.Next()
	}
}

// Rate Limiter en mémoire pour le login (ip -> limiter)
var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

// getVisitorLimiter récupère ou crée un limiteur pour l'adresse IP donnée
func getVisitorLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		// 5 requêtes par seconde, avec un "burst" maximum de 10
		limiter = rate.NewLimiter(5, 10)
		visitors[ip] = limiter
	}

	return limiter
}

// RateLimitLogin protège la route de login contre la force brute
func RateLimitLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitorLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Trop de tentatives de connexion, veuillez patienter."})
			c.Abort()
			return
		}

		c.Next()
	}
}
