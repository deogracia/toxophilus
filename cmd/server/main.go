package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/handlers"
	"github.com/deogracia/toxophilus/internal/logger"
	"github.com/deogracia/toxophilus/internal/middleware"
	"github.com/deogracia/toxophilus/services"
	"github.com/deogracia/toxophilus/static"
	"github.com/deogracia/toxophilus/templates"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// setupRouter configure Gin en fonction de l'environnement
func setupRouter(env string, logFile *os.File) *gin.Engine {
	// Initialisation d'un routeur VIERGE (sans le logger par défaut de Gin)
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// On initialise nos dépôts et handlers avec injection de dépendances
	memberRepo := database.NewGormMemberRepository(database.DB)
	memberHandler := handlers.NewMemberHandler(memberRepo)

	riserRepo := database.NewGormRiserRepository(database.DB)
	limbRepo := database.NewGormLimbRepository(database.DB)
	equipementHandler := handlers.NewEquipementHandler(riserRepo, limbRepo)

	// On attache NOTRE logger, ainsi que le module Recovery (qui évite que le serveur crash en cas de panic)
	r.Use(middleware.SlogLogger(), gin.Recovery())

	// On redirige les petits messages internes de démarrage de Gin vers le même fichier
	if logFile != nil {
		gin.DefaultWriter = io.MultiWriter(os.Stdout, logFile)
	} else {
		gin.DefaultWriter = os.Stdout
	}

	if env == "development" {
		fmt.Println("🚀 Mode DÉVELOPPEMENT (lecture sur disque)")
		templ := template.Must(template.ParseFS(os.DirFS("templates"), "*.html", "partials/*.html"))
		r.SetHTMLTemplate(templ)
		r.StaticFile("/favicon.ico", "static/favicon.ico")
	} else {
		fmt.Println("📦 Mode PRODUCTION (lecture depuis l'exécutable)")
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
		web.GET("/members", memberHandler.GetMembersPage)

		// Page pour modifier un membre
		web.GET("/members/edit/:id", memberHandler.GetMemberEditPage)

		// Page des archives (Membres supprimés)
		web.GET("/members/archives", memberHandler.GetMemberArchivesPage)

		// les routes spécifiques aux matériel
		web.GET("/equipement", equipementHandler.GetEquipementPage)
		web.GET("/equipement/edit/riser/:id", equipementHandler.GetEditRiserPage)
		web.GET("/equipement/edit/limb/:id", equipementHandler.GetEditLimbPage)
		web.GET("/equipement/archives", equipementHandler.GetEquipementArchivesPage)

		// les routes pour les contrats
		web.GET("/contracts", handlers.GetContractsPage)
		web.GET("/contracts/new", handlers.GetNewContractPage)
		web.GET("/contracts/:id", handlers.GetContractDetailsPage)
		web.GET("/contracts/:id/pdf", handlers.DownloadContractPDF)
		web.PUT("/contracts/:id/status", handlers.UpdateContractStatus)

		// les routes pour les settings
		web.GET("/settings", handlers.GetSettingsPage)
		web.POST("/settings/save", handlers.ProcessSettingsSave)

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
	// health check
	r.GET("/health", func(c *gin.Context) {
		// 1. On vérifie si la connexion à la base de données est toujours active
		sqlDB, err := database.DB.DB()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Erreur DB"})
			return
		}

		// 2. On tente un "Ping" réel vers SQLite
		if err := sqlDB.Ping(); err != nil {
			// Si la DB est bloquée, on renvoie une Erreur 500
			// La commande 'curl -f' de l'hébergeur va planter et redémarrer l'application !
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "DB inaccessible"})
			return
		}

		// 3. Tout va bien, on renvoie un simple statut 200 OK
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// 4. Routes Protégées
	api := r.Group("/api")
	api.Use(middleware.AuthRequired())
	{
		// Gestion des membres
		api.GET("/members", memberHandler.ListMembers)
		api.POST("/members", memberHandler.CreateMember)
		api.PUT("/members/:id", memberHandler.UpdateMember)
		api.DELETE("/members/:id", memberHandler.DeleteMember)
		api.GET("/members/:id/export", memberHandler.ExportMemberData)
		api.DELETE("/members/:id/hard", memberHandler.HardDeleteMember)
		api.PUT("/members/:id/reactivate", memberHandler.ReactivateMember)

		// Gestion équipement
		//  Poignée
		api.GET("/risers", equipementHandler.ListRisers)
		api.POST("/risers", equipementHandler.CreateRiser)
		api.PUT("/risers/:id", equipementHandler.UpdateRiser)
		api.DELETE("/risers/:id", equipementHandler.DeleteRiser)
		api.DELETE("/risers/:id/hard", equipementHandler.HardDeleteRiser)
		api.PUT("/risers/:id/reactivate", equipementHandler.ReactivateRiser)

		// Branches
		api.GET("/limbs", equipementHandler.ListLimbs)
		api.POST("/limbs", equipementHandler.CreateLimb)
		api.PUT("/limbs/:id", equipementHandler.UpdateLimb)
		api.DELETE("/limbs/:id", equipementHandler.DeleteLimb)
		api.DELETE("/limbs/:id/hard", equipementHandler.HardDeleteLimb)
		api.PUT("/limbs/:id/reactivate", equipementHandler.ReactivateLimb)

		// Contrats
		api.POST("/contracts", handlers.CreateContract)

	}

	return r
}

func main() {
	errConfig := config.LoadConfig()

	// 1. Initialisation du Logger (Mode Debug = true, Format = "texte" ou "json")
	// En production, tu pourras lire ces valeurs depuis ton config.toml

	logFilePath := viper.GetString("log.file")
	logLevel := viper.GetString("log.level")
	logFormat := viper.GetString("log.format")

	logFile := logger.InitLogger(logFilePath, logLevel, logFormat)
	defer logFile.Close()

	// 2. Verbosité au démarrage avec slog
	slog.Info("🏹 Démarrage de Toxophilus",
		slog.String("version", config.AppVersion),
		slog.String("environnement", viper.GetString("app.env")),
	)

	if errConfig != nil {
		slog.Info("ℹ️ Configuration : utilisation des variables d'environnement et/ou valeurs par défaut",
			slog.String("detail", errConfig.Error()),
		)
	} else {
		slog.Info("✅ Configuration : fichier config.toml chargé avec succès")
	}

	if viper.GetString("app.secret_key") == "" {
		log.Fatal("🛑 ERREUR FATALE : la clé de configuration 'app.secret_key' est manquante. Le serveur refuse de démarrer pour des raisons de sécurité.")
	}

	slog.Info("Connexion à la base de données "+viper.GetString("database.driver"), slog.String("DSN", viper.GetString("database.dsn")))
	database.Connect()
	services.InitDefaultSettings()

	// 0. Gestion de l'environnement (Prduction par défaut)
	env := os.Getenv("APP_ENV")
	switch env {
	case "", "development", "production":
		log.Println("APP_ENV a une valeur autorisée!")
	default:
		// #nosec G706 -- l'application s'arrête immédiatement.
		log.Fatalf("🛑 ERREUR FATALE : Environement de démarage %q non pris en charge. L'application s'arrête.", env)
	}

	if env == "" {
		env = "production"
	}

	r := setupRouter(env, logFile)

	port := viper.GetString("app.port")
	if port == "" {
		port = "8080"
	}

	// On formate le port pour correspondre à `:PORT`
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	// On configure le serveur HTTP natif de Go avec le routeur Gin
	srv := &http.Server{
		Addr:              port,
		ReadHeaderTimeout: 5 * time.Second, // protection anti Slowloris
		Handler:           r,
	}

	// On lance le serveur dans une Goroutine (en arrière-plan) pour ne pas bloquer la suite du code
	go func() {
		slog.Info("🚀 Serveur en écoute", slog.String("port", port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("🛑 Erreur critique du serveur : %v", err)
		}
	}()

	// On crée un canal pour écouter les signaux d'arrêt du système d'exploitation
	quit := make(chan os.Signal, 1)
	// kill (sans paramètre) par défaut envoie syscal.SIGTERM
	// kill -2 est syscall.SIGINT (ce que fait le CTRL+C)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Le code se met en pause ICI et attend qu'un signal soit reçu dans le canal
	<-quit
	slog.Info("🛑 Signal d'arrêt reçu. Fermeture gracieuse en cours...")

	// On donne 5 secondes au serveur pour finir ce qu'il est en train de faire
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("❌ Le serveur a été forcé de s'arrêter", slog.Any("erreur", err))
	}

	// On ferme proprement la connexion à la base de données SQLite
	sqlDB, err := database.DB.DB()
	if err == nil {
		_ = sqlDB.Close() // On ignore l'erreur car le serveur est en train de s'éteindre
		slog.Info("💾 Connexion à la base de données fermée.")
	}

	slog.Info("👋 Toxophilus s'est arrêté proprement. À bientôt !")
}
