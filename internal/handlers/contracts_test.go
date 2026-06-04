package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupContractTestDB() *gin.Engine {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Erreur lors de l'ouverture de la base de test: " + err.Error())
	}
	database.DB = db

	// Migration des tables nécessaires
	err = database.DB.AutoMigrate(&models.Member{}, &models.Riser{}, &models.Limb{}, &models.Contract{})
	if err != nil {
		panic("Erreur lors de la migration de test: " + err.Error())
	}

	r := gin.Default()

	tmpl := template.Must(template.New("contract_details.html").Parse("<h1>Template de test</h1>"))
	r.SetHTMLTemplate(tmpl)
	return r
}

func TestCreateContractAndUpdateStock(t *testing.T) {
	r := setupContractTestDB()

	// Déclaration de la route à tester
	r.POST("/api/contracts", CreateContract)

	// ==========================================
	// 1. PRÉPARATION DES DONNÉES EN BASE
	// ==========================================
	member := models.Member{CodeAdherent: "FR-2026", Nom: "Test", Prenom: "Archer"}
	riser := models.Riser{NumeroSerie: "R-TEST-1", Disponibilite: true}
	limb := models.Limb{NumeroSerie: "L-TEST-1", Disponibilite: true}

	database.DB.Create(&member)
	database.DB.Create(&riser)
	database.DB.Create(&limb)

	// ==========================================
	// 2. SIMULATION DE L'ENVOI DU FORMULAIRE
	// ==========================================
	formData := url.Values{}
	formData.Set("member_id", "1")
	formData.Set("date_debut", "2026-05-01")
	formData.Set("date_fin", "2026-05-31")
	formData.Set("riser_id", "1")
	formData.Set("limb_id", "1")
	formData.Set("montant_location", "150.0")
	formData.Set("montant_caution", "300.0")
	formData.Set("etat_paiement", "En attente")

	// Création de la requête au format standard (application/x-www-form-urlencoded)
	req, _ := http.NewRequest("POST", "/api/contracts", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// ==========================================
	// 3. VÉRIFICATIONS (ASSERTS)
	// ==========================================

	// A. Le code HTTP doit être 201 (Created)
	if w.Code != http.StatusCreated {
		t.Fatalf("❌ Code attendu 201, obtenu %d", w.Code)
	}

	// B. L'ordre de redirection HTMX doit être présent
	if w.Header().Get("HX-Redirect") != "/contracts" {
		t.Errorf("❌ Redirection HX-Redirect manquante ou incorrecte")
	}

	// C. Le contrat doit exister en base
	var contract models.Contract
	database.DB.First(&contract)
	if contract.ID == 0 {
		t.Fatalf("❌ Le contrat n'a pas été enregistré en base")
	}

	// D. VÉRIFICATION CRUCIALE : Les stocks ont-ils été mis à jour ?
	var updatedRiser models.Riser
	database.DB.First(&updatedRiser, riser.ID)
	if updatedRiser.Disponibilite != false {
		t.Errorf("❌ La disponibilité de la poignée aurait dû passer à false")
	}

	var updatedLimb models.Limb
	database.DB.First(&updatedLimb, limb.ID)
	if updatedLimb.Disponibilite != false {
		t.Errorf("❌ La disponibilité des branches aurait dû passer à false")
	}
}

func TestGetContractDetailsPage(t *testing.T) {
	r := setupContractTestDB()

	// Déclaration de la route à tester
	r.GET("/contracts/:id", GetContractDetailsPage)

	// ==========================================
	// 1. PRÉPARATION DES DONNÉES EN BASE
	// ==========================================
	member := models.Member{CodeAdherent: "FR-2026-DETAILS", Nom: "Test", Prenom: "Details"}
	database.DB.Create(&member)

	// Création d'un vrai contrat avec l'ID 1
	contract := models.Contract{
		MemberID:        member.ID,
		MontantLocation: 150.0,
		EtatPaiement:    "Payé",
	}
	database.DB.Create(&contract)

	// ==========================================
	// 2. TEST A : Le contrat existe (Succès 200)
	// ==========================================
	reqFound, _ := http.NewRequest("GET", "/contracts/1", nil)
	wFound := httptest.NewRecorder()
	r.ServeHTTP(wFound, reqFound)

	if wFound.Code != http.StatusOK {
		t.Errorf("❌ Code attendu 200 pour un contrat existant, obtenu %d", wFound.Code)
	}

	// ==========================================
	// 3. TEST B : Le contrat n'existe pas (Erreur 404)
	// ==========================================
	reqNotFound, _ := http.NewRequest("GET", "/contracts/99999", nil)
	wNotFound := httptest.NewRecorder()
	r.ServeHTTP(wNotFound, reqNotFound)

	if wNotFound.Code != http.StatusNotFound {
		t.Errorf("❌ Code attendu 404 pour un contrat inexistant, obtenu %d", wNotFound.Code)
	}

	// Vérification supplémentaire : le message d'erreur est bien présent dans la réponse
	if !strings.Contains(wNotFound.Body.String(), "Erreur: le contrat 99999 n'existe pas!") {
		t.Errorf("❌ Le corps de la réponse ne contient pas le message d'erreur 404 attendu")
	}
}

func TestUpdateContractStatus(t *testing.T) {
	// 1. Préparation de la base de données de test
	r := setupContractTestDB()

	// Création d'une poignée et de branches factices considérées comme louées (non disponibles)
	riser := models.Riser{Marque: "Hoyt Test", Disponibilite: false}
	limb := models.Limb{Marque: "Uukha Test", Disponibilite: false}
	database.DB.Create(&riser)
	database.DB.Create(&limb)

	// Création d'un contrat factice (Actif) lié à ce matériel
	contract := models.Contract{
		Statut:       "Actif",
		EtatPaiement: "En attente",
		RiserID:      &riser.ID,
		LimbID:       &limb.ID,
	}
	database.DB.Create(&contract)

	// 2. Ajout de la route à tester
	r.PUT("/contracts/:id/status", UpdateContractStatus)

	// 3. Préparation des données du formulaire (Simulation de HTMX)
	formData := url.Values{}
	formData.Set("statut", "Terminé")
	formData.Set("etat_paiement", "Payé")
	formData.Set("mode_paiement", "CB")
	formData.Set("recu_signe", "true")

	// 4. Exécution de la requête HTTP PUT
	w := httptest.NewRecorder()
	urlPath := "/contracts/" + strconv.Itoa(int(contract.ID)) + "/status"
	req, _ := http.NewRequest(http.MethodPut, urlPath, strings.NewReader(formData.Encode()))
	// Indispensable pour que Gin comprenne qu'il s'agit d'un formulaire (c.PostForm)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	r.ServeHTTP(w, req)

	// ==========================================
	// 5. VÉRIFICATIONS (Assertions)
	// ==========================================

	// A. Vérifier le code HTTP et le header spécifique à HTMX
	if w.Code != http.StatusOK {
		t.Errorf("❌ Code HTTP attendu 200, obtenu %d", w.Code)
	}
	if w.Header().Get("HX-Refresh") != "true" {
		t.Errorf("❌ Header HTMX 'HX-Refresh' manquant")
	}

	// B. Vérifier que le contrat a bien été mis à jour en base
	var updatedContract models.Contract
	database.DB.First(&updatedContract, contract.ID)

	if updatedContract.Statut != "Terminé" {
		t.Errorf("❌ Statut attendu 'Terminé', obtenu '%s'", updatedContract.Statut)
	}
	if updatedContract.ModePaiement != "CB" {
		t.Errorf("❌ Mode de paiement attendu 'CB', obtenu '%s'", updatedContract.ModePaiement)
	}
	if !updatedContract.RecuSigne {
		t.Errorf("❌ RecuSigne attendu true, obtenu false")
	}

	// C. Vérifier la libération logique du matériel (Le plus important !)
	var updatedRiser models.Riser
	database.DB.First(&updatedRiser, riser.ID)
	if !updatedRiser.Disponibilite {
		t.Errorf("❌ La poignée devrait être redevenue disponible")
	}

	var updatedLimb models.Limb
	database.DB.First(&updatedLimb, limb.ID)
	if !updatedLimb.Disponibilite {
		t.Errorf("❌ Les branches devraient être redevenues disponibles")
	}
}
