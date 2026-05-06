package middleware

import (
	"net/http"

	"github.com/deogracia/toxophilus/internal/auth"
	"github.com/gin-gonic/gin"
)

// AuthRequired protège les routes en vérifiant la présence et la validité du cookie JWT
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Récupérer le token depuis le cookie
		tokenString, err := c.Cookie("toxo_session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Accès restreint, veuillez vous connecter"})
			c.Abort()
			return
		}

		// 2. Valider le token avec notre fonction métier
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			// Le token est invalide, expiré, ou a été trafiqué
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session invalide ou expirée, veuillez vous reconnecter"})
			c.Abort()
			return
		}

		// 3. Transmettre l'ID de l'utilisateur aux prochains gestionnaires (handlers)
		// C'est extrêmement utile si une route a besoin de savoir QUI fait la modification
		c.Set("userID", claims.UserID)

		// 4. Tout est bon, on laisse passer la requête
		c.Next()
	}
}
