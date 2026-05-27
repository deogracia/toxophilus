package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/deogracia/toxophilus/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetContractsPage affiche la liste de tous les contrats de location
func GetContractsPage(c *gin.Context) {
	var contracts []models.Contract

	// Preload charge automatiquement les données associées via les clés étrangères.
	// C'est indispensable pour que {{ .Member.Prenom }} ne soit pas vide dans le HTML.
	err := database.DB.Preload("Member").Preload("Riser").Preload("Limb").Find(&contracts).Error
	if err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors du chargement des contrats : %v", err)
		return
	}

	// 🔍 BOUCLE DE LOGS DE DÉBOGAGE
	for i, contract := range contracts {
		log.Printf("=================GetContractPage - debug ========================")
		log.Printf("🔍 [DEBUG] CONTRAT INTERNE N°%d (ID Base: %d)", i+1, contract.ID)
		if contract.RiserID == nil {
			log.Printf(" -> RiserID est strictement NIL (NULL en BDD)")
		} else {
			log.Printf(" -> RiserID pointe vers l'adresse : %v (Valeur contenue: %d)", contract.RiserID, *contract.RiserID)
		}
		log.Printf(" -> Structure Riser interne : ID=%d, Numéro de Série=%q", contract.Riser.ID, contract.Riser.NumeroSerie)
	}
	log.Printf("==================Fin GetContractPage - debug=======================")

	c.HTML(http.StatusOK, "contracts.html", gin.H{
		"titre":     "Contrats de Location - Club Toxophilus",
		"active":    "contracts", // Permet d'allumer le bon onglet dans la nav
		"Contracts": contracts,
	})
}

// GetNewContractPage affiche le formulaire de création d'un contrat
func GetNewContractPage(c *gin.Context) {
	var members []models.Member
	var risers []models.Riser
	var limbs []models.Limb

	// On récupère les membres triés par ordre alphabétique
	database.DB.Order("nom ASC").Find(&members)

	// On récupère le matériel disponible (GORM ignore automatiquement ceux qui sont dans les archives)
	database.DB.Where("disponibilite = ?", true).Find(&risers)
	database.DB.Where("disponibilite = ?", true).Find(&limbs)

	c.HTML(http.StatusOK, "contract_new.html", gin.H{
		"titre":   "Nouveau Contrat - Club Toxophilus",
		"active":  "contracts",
		"Members": members,
		"Risers":  risers,
		"Limbs":   limbs,
	})
}

// CreateContract réceptionne le formulaire standard et crée le contrat en base
func CreateContract(c *gin.Context) {
	// Gin va utiliser les tags 'form' pour lier automatiquement le texte HTML aux bons types Go !
	var input struct {
		MemberID        uint    `form:"member_id" binding:"required"`
		DateDebut       string  `form:"date_debut" binding:"required"`
		DateFin         string  `form:"date_fin" binding:"required"`
		RiserID         *uint   `form:"riser_id"` // Si vide, Gin le laisse à nil automatiquement
		LimbID          *uint   `form:"limb_id"`  // Si vide, Gin le laisse à nil automatiquement
		Accessoires     string  `form:"accessoires"`
		Commentaire     string  `form:"commentaire"`
		MontantLocation float64 `form:"montant_location"`
		MontantCaution  float64 `form:"montant_caution"`
		EtatPaiement    string  `form:"etat_paiement"`
		ModePaiement    string  `form:"mode_paiement"`
	}

	// ShouldBind au lieu de ShouldBindJSON lit directement le Form Data standard
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Veuillez vérifier les champs du formulaire."})
		return
	}

	// 🔴 BLOC DE DÉBOGAGE : Inspection brute du formulaire reçu
	log.Println("============ GIN BINDING DEBUG ============")
	if input.RiserID == nil {
		log.Println("🔍 RiserID reçu du HTML : est strictement NIL (Pointeur nul)")
	} else {
		log.Printf("🔍 RiserID reçu du HTML : n'est PAS NIL ! Adresse: %v, Valeur pointée: %d", input.RiserID, *input.RiserID)
	}

	if input.LimbID == nil {
		log.Println("🔍 LimbID reçu du HTML  : est strictement NIL (Pointeur nul)")
	} else {
		log.Printf("🔍 LimbID reçu du HTML  : n'est PAS NIL ! Adresse: %v, Valeur pointée: %d", input.LimbID, *input.LimbID)
	}
	log.Println("==============Fin Gin Binding Debug=============================")

	// 🛠️ SÉCURITÉ : Si Gin a lié une chaîne vide vers un pointeur de 0, on le remet à nil
	if input.RiserID != nil && *input.RiserID == 0 {
		input.RiserID = nil
	}
	if input.LimbID != nil && *input.LimbID == 0 {
		input.LimbID = nil
	}

	// Conversion des dates
	debut, err := time.Parse("2006-01-02", input.DateDebut)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format de date de début invalide"})
		return
	}
	fin, err := time.Parse("2006-01-02", input.DateFin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format de date de fin invalide"})
		return
	}

	// Initialisation du contrat avec les données directes
	contract := models.Contract{
		MemberID:        input.MemberID,
		DateDebut:       debut,
		DateFin:         fin,
		RiserID:         input.RiserID,
		LimbID:          input.LimbID,
		Accessoires:     input.Accessoires,
		Commentaire:     input.Commentaire,
		MontantLocation: input.MontantLocation,
		MontantCaution:  input.MontantCaution,
		EtatPaiement:    input.EtatPaiement,
		ModePaiement:    input.ModePaiement,
	}

	if err := database.DB.Omit("Riser", "Limb").Create(&contract).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'enregistrement du contrat"})
		return
	}

	// MISE À JOUR DES STOCKS (Disponibilité)
	// Si une poignée a été louée, on la marque comme indisponible
	if input.RiserID != nil {
		database.DB.Model(&models.Riser{}).Where("id = ?", input.RiserID).Update("disponibilite", false)
	}
	// Si des branches ont été louées, on les marque comme indisponibles
	if input.LimbID != nil {
		database.DB.Model(&models.Limb{}).Where("id = ?", input.LimbID).Update("disponibilite", false)
	}
	c.Header("HX-Redirect", "/contracts")
	c.Status(http.StatusCreated)
}

// GetContractDetailsPage affiche le récapitulatif complet d'un contrat spécifique
func GetContractDetailsPage(c *gin.Context) {
	id := c.Param("id")
	var contract models.Contract

	// On précharge toutes les relations pour avoir les infos complètes
	err := database.DB.Preload("Member").Preload("Riser").Preload("Limb").First(&contract, id).Error
	if err != nil {
		// Enregistrement non trouvé
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Erreur: le contrat %s n'existe pas!", id)
			return
		}
		// pour toute autre erreur, on
		c.String(http.StatusInternalServerError, "Erreur lors du chargement du contrat %s! Erreur: %v", id, err.Error())
		return
	}

	c.HTML(http.StatusOK, "contract_details.html", gin.H{
		"titre":    "Détails du contrat #" + id,
		"active":   "contracts",
		"Contract": contract,
	})
}

// DownloadContractPDF génère et envoie le PDF au navigateur
func DownloadContractPDF(c *gin.Context) {
	id := c.Param("id")
	var contract models.Contract

	// 1. Récupération du contrat
	if err := database.DB.Preload("Member").Preload("Riser").Preload("Limb").First(&contract, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Erreur 404 : Le contrat n'existe pas.")
			return
		}
		c.String(http.StatusInternalServerError, "Erreur serveur : %v", err)
		return
	}

	// 2. Récupération des réglages dynamiques pour l'Open Source
	var settingsList []models.Setting
	database.DB.Find(&settingsList)

	settingsMap := make(map[string]string)
	for _, s := range settingsList {
		settingsMap[s.Cle] = s.Valeur
		log.Println("contracts.go - boucle de transformation. Clé -> " + s.Cle + " Valeur -> " + s.Valeur)
	}

	// 3. Appel du service de génération avec les settings
	filename, err := services.GenerateContractPDF(contract, settingsMap)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la création du PDF : %v", err)
		return
	}

	// 4. Téléchargement
	c.FileAttachment(filename, filename)
}
