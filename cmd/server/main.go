package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/handlers"
	"github.com/deogracia/toxophilus/internal/middleware"
	"github.com/deogracia/toxophilus/models"
	"github.com/deogracia/toxophilus/services"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	config.LoadConfig()

	if viper.GetString("app.secret_key") == "" {
		log.Fatal("🛑 ERREUR FATALE : la clé de configuration 'app.secret_key' est manquante. Le serveur refuse de démarrer pour des raisons de sécurité.")
	}

	database.Connect()
	services.InitDefaultSettings()

	r := gin.Default()
	// 0. CHARGEMENT DES TEMPLATES HTML ---
	// On charge d'abord les pages principales
	templ := template.Must(template.ParseGlob("templates/*.html"))
	// On y ajoute (fusionne) les fragments du sous-dossier
	template.Must(templ.ParseGlob("templates/partials/*.html"))

	r.SetHTMLTemplate(templ)

	// 1. Le Gatekeeper : Force le setup si la base est vierge
	r.Use(middleware.EnsureSetup())

	// ==========================================
	// 🌐 PARTIE FRONT-END (HTML)
	// ==========================================

	// Route publique pour afficher le formulaire de connexion
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{"titre": "Connexion - Toxophilus"})
	})

	// Groupe pour les pages HTML privées (le Dashboard)
	web := r.Group("/")
	web.Use(middleware.AuthRequired()) // On recycle notre middleware JWT !
	{
		web.GET("/", func(c *gin.Context) {

			c.HTML(http.StatusOK, "index.html", gin.H{
				"titre":  "Dashboard - Toxophilus",
				"active": "accueil",
			})
		})
		web.GET("/members", func(c *gin.Context) {
			// On récupère la liste des membres
			var members []models.Member
			database.DB.Find(&members)

			c.HTML(http.StatusOK, "members.html", gin.H{
				"titre":   "Gestion des membres - Toxophilus",
				"membres": members,
				"active":  "members",
			})
		})
		// Page pour modifier un membre
		web.GET("/members/edit/:id", func(c *gin.Context) {
			id := c.Param("id")
			var member models.Member

			// On cherche le membre. S'il n'existe pas, on redirige vers la liste
			if err := database.DB.First(&member, id).Error; err != nil {
				c.Redirect(http.StatusTemporaryRedirect, "/members")
				return
			}

			// On affiche la nouvelle page avec les données du membre pré-chargées
			c.HTML(http.StatusOK, "member_edit.html", gin.H{
				"titre":  "Modifier - Toxophilus",
				"active": "membres", // Pour garder le menu "Membres" allumé !
				"member": member,    // On passe les données de l'archer
			})
		})
		// Page des archives (Membres supprimés)
		web.GET("/members/archives", func(c *gin.Context) {
			var archivedMembers []models.Member
			// On va chercher uniquement les membres qui ont une date de suppression
			database.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&archivedMembers)

			c.HTML(http.StatusOK, "member_archives.html", gin.H{
				"titre":   "Les alumnis - Toxophilus",
				"active":  "membres", // On garde l'onglet Membres allumé
				"membres": archivedMembers,
			})
		})

		// les routes spécifiques aux matériel
		web.GET("/materiel", handlers.GetMaterielPage)
	}

	// 2. Routes de Configuration Initiale
	r.GET("/setup", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Veuillez envoyer un POST sur /setup/process avec admin_email et admin_password"})
	})
	r.POST("/setup/process", handlers.ProcessSetup)

	// ==========================================
	// ⚙️ PARTIE BACK-END (API REST JSON)
	// ==========================================

	// 3. Routes Publiques
	r.POST("/api/login", handlers.LoginHandler)
	r.POST("/api/logout", handlers.LogoutHandler)

	// 4. Routes Protégées
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		// 	api.GET("/risers", handlers.ListRisers)

		// on teste que ça se goupille bien.
		api.GET("/me", func(ctx *gin.Context) {
			userID, _ := ctx.Get("userID")
			ctx.JSON(200, gin.H{"message": "Accès Autorisé", "ton_id_utilisateur": userID})

		})

		// Gestion des membres
		api.GET("/members", handlers.ListMembers)
		api.POST("/members", handlers.CreateMember)
		api.PUT("/members/:id", handlers.UpdateMember)
		api.DELETE("/members/:id", handlers.DeleteMember)
		api.GET("/members/:id/export", handlers.ExportMemberData)
		api.DELETE("/members/:id/hard", handlers.HardDeleteMember)
		api.PUT("/members/:id/reactivate", handlers.ReactivateMember)

		// Gestion équipement
		//  Poignée
		api.GET("/risers", handlers.ListRisers)
		api.POST("/risers", handlers.CreateRiser)
		api.PUT("/risers/:id", handlers.UpdateRiser)
		api.DELETE("/risers/:id", handlers.DeleteRiser)

		// BRanches
		api.GET("/limbs", handlers.ListLimbs)
		api.POST("/limbs", handlers.CreateLimb)
		api.PUT("/limbs/:id", handlers.UpdateLimb)
		api.DELETE("/limbs/:id", handlers.DeleteLimb)

	}

	port := viper.GetString("app.port")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Démarrage du serveur sur http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Erreur lors du lancement du serveur: %v", err)
	}
}
