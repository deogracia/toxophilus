package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/handlers"
	"github.com/deogracia/toxophilus/internal/middleware"
	"github.com/deogracia/toxophilus/services"
	"github.com/deogracia/toxophilus/static"
	"github.com/deogracia/toxophilus/templates"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// setupRouter configure Gin en fonction de l'environnement
func setupRouter(env string) *gin.Engine {
	r := gin.Default()

	if env == "development" {
		fmt.Println("🚀 Mode DÉVELOPPEMENT (lecture sur disque)")
		templ := template.Must(template.ParseFS(os.DirFS("templates"), "*.html", "partials/*.html"))
		r.SetHTMLTemplate(templ)
		r.StaticFile("/favicon.ico", "static/favicon.ico")
	} else {
		fmt.Println("📦 Mode PRODUCTION (lecture depuis l'exécutable)")
		gin.SetMode(gin.ReleaseMode)
		templ := template.Must(template.ParseFS(templates.TemplateFS, "*.html", "partials/*.html"))
		r.SetHTMLTemplate(templ)
		r.StaticFileFS("/favicon.ico", "favicon.ico", http.FS(static.StaticFS))
	}

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
		web.GET("/", handlers.GetDashboardPage)

		// List des membres
		web.GET("/members", handlers.GetMembersPage)

		// Page pour modifier un membre
		web.GET("/members/edit/:id", handlers.GetMemberEditPage)

		// Page des archives (Membres supprimés)
		web.GET("/members/archives", handlers.GetMemberArchivesPage)

		// les routes spécifiques aux matériel
		web.GET("/equipement", handlers.GetEquipementPage)
		web.GET("/equipement/edit/riser/:id", handlers.GetEditRiserPage)
		web.GET("/equipement/edit/limb/:id", handlers.GetEditLimbPage)

	}

	// 2. Routes de Configuration Initiale
	r.GET("/setup", handlers.GetSetupPage)
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

		// Branches
		api.GET("/limbs", handlers.ListLimbs)
		api.POST("/limbs", handlers.CreateLimb)
		api.PUT("/limbs/:id", handlers.UpdateLimb)
		api.DELETE("/limbs/:id", handlers.DeleteLimb)

	}

	return r
}

func main() {
	config.LoadConfig()

	if viper.GetString("app.secret_key") == "" {
		log.Fatal("🛑 ERREUR FATALE : la clé de configuration 'app.secret_key' est manquante. Le serveur refuse de démarrer pour des raisons de sécurité.")
	}

	database.Connect()
	services.InitDefaultSettings()

	// 0. Gestion de l'environnement (Prduction par défaut)
	env := os.Getenv("APP_ENV")
	switch env {
	case "", "development", "production":
		log.Println("APP_ENV a une valeur autorisée!")
	default:
		log.Fatalf("🛑 ERREUR FATALE : Environement de démarage %s non pris en charge. L'application s'arrête.", env)
	}

	if env == "" {
		env = "production"
	}

	r := setupRouter(env)

	port := viper.GetString("app.port")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Démarrage du serveur sur http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Erreur lors du lancement du serveur: %v", err)
	}
}
