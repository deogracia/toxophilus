package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// SlogLogger remplace le logger par défaut de Gin pour utiliser log/slog
func SlogLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// On note l'heure de début
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// On laisse Gin traiter la requête (appeler les autres handlers)
		c.Next()

		// On calcule la durée une fois la requête terminée
		cost := time.Since(start)
		status := c.Writer.Status()

		// On ignore les requêtes vers les fichiers statiques pour ne pas polluer les logs (optionnel)
		if len(path) >= 8 && path[:8] == "/static/" {
			return
		}

		// On envoie le tout à slog (qui gérera le fichier, la console, et le format text/json)
		if status >= 400 {
			// S'il y a une erreur HTTP (404, 500...), on loggue en niveau Error ou Warn
			slog.Error("Requête HTTP échouée",
				slog.Int("status", status),
				slog.String("methode", c.Request.Method),
				slog.String("chemin", path),
				slog.String("ip", c.ClientIP()),
				slog.Duration("duree", cost),
			)
		} else {
			// Sinon, en niveau Info
			slog.Info("Requête HTTP",
				slog.Int("status", status),
				slog.String("methode", c.Request.Method),
				slog.String("chemin", path),
				slog.String("requete", query),
				slog.String("ip", c.ClientIP()),
				slog.Duration("duree", cost),
			)
		}
	}
}
