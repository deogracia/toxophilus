package main

import (
	"fmt"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/services"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/spf13/viper"
)

// E2ETestEnv regroupe les éléments nécessaires pour les tests d'interface
type E2ETestEnv struct {
	Server  *httptest.Server
	Browser *rod.Browser
}

// setupE2EEnv initialise l'environnement complet (DB, Serveur Web, Navigateur)
func setupE2EEnv(t *testing.T) *E2ETestEnv {
	t.Helper()

	// Configuration
	viper.Set("app.env", "production") // Pour utiliser les templates embedded
	viper.Set("database.driver", "sqlite")
	// On utilise une DB partagée en mémoire pour la durée du test
	viper.Set("database.dsn", "file::memory:?cache=shared")
	viper.Set("app.secret_key", "cle_super_secrete_pour_les_tests_qui_fait_plus_de_32_caracteres_!")
	viper.Set("app.data_dir", os.TempDir())

	// Init DB
	database.Connect()

	// S'assurer que la base est propre pour chaque test (si les tests s'exécutent séquentiellement sur le même cache shared)
	database.DB.Exec("DELETE FROM users")
	database.DB.Exec("DELETE FROM members")
	database.DB.Exec("DELETE FROM settings")

	services.InitDefaultSettings()

	// Init Serveur
	r := setupRouter("production", nil)
	ts := httptest.NewServer(r)

	// Init Navigateur
	// On désactive la sandbox pour que Chromium puisse tourner dans Docker / GitHub Actions sans droits root spéciaux.
	l := launcher.New().NoSandbox(true)

	// On désactive Leakless uniquement sur Windows pour éviter les faux positifs d'antivirus.
	// Sur Linux (CI) et macOS, on le garde pour éviter les processus zombies en cas de crash violent.
	if runtime.GOOS == "windows" {
		l = l.Leakless(false)
	}

	u := l.MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	// Cleanup automatique à la fin du test
	t.Cleanup(func() {
		browser.MustClose()
		ts.Close()
	})

	return &E2ETestEnv{
		Server:  ts,
		Browser: browser,
	}
}

// CreateAdmin injecte un administrateur directement en base de données pour contourner le Setup
func (env *E2ETestEnv) CreateAdmin(email, password string) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(fmt.Sprintf("Impossible de hasher le mot de passe pour le test: %v", err))
	}
	database.DB.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, string(hashedPassword))
}

// ============================================================================
// TESTS
// ============================================================================

func TestE2E_SetupPage(t *testing.T) {
	env := setupE2EEnv(t)

	// On accède à l'accueil, ce qui doit nous rediriger vers /setup car il n'y a pas d'admin
	page := env.Browser.MustPage(env.Server.URL + "/")
	page.MustWaitLoad()
	time.Sleep(200 * time.Millisecond)

	bodyText := page.MustElement("body").MustText()

	if !strings.Contains(bodyText, "Configuration initiale") {
		t.Errorf("Le texte 'Configuration initiale' est introuvable. Texte obtenu : %s", bodyText)
	}
	if !strings.Contains(bodyText, "Créer l'administrateur") {
		t.Errorf("Le bouton 'Créer l'administrateur' est introuvable.")
	}
}

func TestE2E_LoginPage(t *testing.T) {
	env := setupE2EEnv(t)
	env.CreateAdmin("admin@toxophilus.local", "password123")

	// Si un admin existe, aller sur / redirige vers le login quand on n'est pas authentifié
	page := env.Browser.MustPage(env.Server.URL + "/")
	page.MustWaitLoad()
	time.Sleep(200 * time.Millisecond)

	bodyText := page.MustElement("body").MustText()

	if !strings.Contains(bodyText, "Veuillez vous connecter pour accéder à l'interface") {
		t.Errorf("Texte de connexion introuvable. On devrait être sur /login.")
	}
	if !strings.Contains(bodyText, "Se connecter") {
		t.Errorf("Le bouton 'Se connecter' est introuvable.")
	}
}

func TestE2E_DashboardAndMembers(t *testing.T) {
	env := setupE2EEnv(t)

	// 1. On crée un admin
	env.CreateAdmin("admin@toxophilus.local", "password123")

	// 2. On va sur la page de login
	page := env.Browser.MustPage(env.Server.URL + "/login")
	page.MustWaitLoad()

	// 3. On remplit le formulaire et on clique
	page.MustElement("#email").MustInput("admin@toxophilus.local")
	page.MustElement("#password").MustInput("password123")
	page.MustElement("button[type=submit]").MustClick()

	// 4. On attend d'être redirigé vers le Dashboard
	page.MustWaitStable()
	time.Sleep(500 * time.Millisecond) // Laisse le temps à l'interface de s'afficher

	bodyText := page.MustElement("body").MustText()
	if !strings.Contains(bodyText, "Tableau de bord") {
		t.Errorf("Impossible d'accéder au tableau de bord. Texte obtenu : %s", bodyText)
	}

	// 5. On teste la navigation vers la page des membres via le menu (s'il s'appelle "Membres" ou l'URL directement)
	page.MustNavigate(env.Server.URL + "/members")
	page.MustWaitStable()
	time.Sleep(200 * time.Millisecond)

	memberBodyText := page.MustElement("body").MustText()
	if !strings.Contains(memberBodyText, "Nouveau Membre") && !strings.Contains(memberBodyText, "Membres") {
		t.Errorf("Impossible d'afficher la page des membres.")
	}
}
