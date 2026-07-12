package handlers

import (
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/models"
	"github.com/deogracia/toxophilus/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ContractHandler gère les requêtes HTTP pour les contrats de location.
type ContractHandler struct {
	contractRepo models.ContractRepository
	memberRepo   models.MemberRepository
	riserRepo    models.RiserRepository
	limbRepo     models.LimbRepository
	settingRepo  models.SettingRepository
}

// NewContractHandler crée une nouvelle instance de ContractHandler.
func NewContractHandler(
	contractRepo models.ContractRepository,
	memberRepo models.MemberRepository,
	riserRepo models.RiserRepository,
	limbRepo models.LimbRepository,
	settingRepo models.SettingRepository,
) *ContractHandler {
	return &ContractHandler{
		contractRepo: contractRepo,
		memberRepo:   memberRepo,
		riserRepo:    riserRepo,
		limbRepo:     limbRepo,
		settingRepo:  settingRepo,
	}
}

// GetContractsPage affiche la liste de tous les contrats de location
func (h *ContractHandler) GetContractsPage(c *gin.Context) {
	contracts, err := h.contractRepo.GetAll()
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
		"Version":   config.AppVersion,
	})
}

// GetNewContractPage affiche le formulaire de création d'un contrat
func (h *ContractHandler) GetNewContractPage(c *gin.Context) {
	members, err := h.memberRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors du chargement des membres : %v", err)
		return
	}

	risers, err := h.riserRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors du chargement des poignées : %v", err)
		return
	}

	limbs, err := h.limbRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors du chargement des branches : %v", err)
		return
	}

	// Filtrer manuellement les disponibles pour la création de contrat
	var availableRisers []models.Riser
	for _, r := range risers {
		if r.Disponibilite {
			availableRisers = append(availableRisers, r)
		}
	}

	var availableLimbs []models.Limb
	for _, l := range limbs {
		if l.Disponibilite {
			availableLimbs = append(availableLimbs, l)
		}
	}

	// Récupération des réglages de caution et loyer
	cautionSetting, _ := h.settingRepo.GetByKey("montant_caution")
	loyerSetting, _ := h.settingRepo.GetByKey("montant_loyer")

	cautionVal := "0"
	if cautionSetting != nil {
		cautionVal = cautionSetting.Valeur
	}
	loyerVal := "0"
	if loyerSetting != nil {
		loyerVal = loyerSetting.Valeur
	}

	c.HTML(http.StatusOK, "contract_new.html", gin.H{
		"titre":   "Nouveau Contrat - Club Toxophilus",
		"active":  "contracts",
		"Members": members,
		"Risers":  availableRisers,
		"Limbs":   availableLimbs,
		"Caution": cautionVal,
		"Loyer":   loyerVal,
		"Version": config.AppVersion,
	})
}

// CreateContract réceptionne le formulaire standard et crée le contrat en base
func (h *ContractHandler) CreateContract(c *gin.Context) {
	var input struct {
		MemberID        uint    `form:"member_id" binding:"required"`
		DateDebut       string  `form:"date_debut" binding:"required"`
		DateFin         string  `form:"date_fin" binding:"required"`
		RiserID         *uint   `form:"riser_id"`
		LimbID          *uint   `form:"limb_id"`
		Accessoires     string  `form:"accessoires"`
		Commentaire     string  `form:"commentaire"`
		MontantLocation float64 `form:"montant_location"`
		MontantCaution  float64 `form:"montant_caution"`
		EtatPaiement    string  `form:"etat_paiement"`
		ModePaiement    string  `form:"mode_paiement"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Veuillez vérifier les champs du formulaire."})
		return
	}

	// 🔴 BLOC DE DÉBOGAGE
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

	if input.RiserID != nil && *input.RiserID == 0 {
		input.RiserID = nil
	}
	if input.LimbID != nil && *input.LimbID == 0 {
		input.LimbID = nil
	}

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

	if err := h.contractRepo.Create(&contract); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de l'enregistrement du contrat"})
		return
	}

	// MISE À JOUR DES STOCKS (Disponibilité) via les repositories d'équipement
	if input.RiserID != nil {
		riser, err := h.riserRepo.GetByID(*input.RiserID)
		if err == nil {
			riser.Disponibilite = false
			_ = h.riserRepo.Update(riser)
		}
	}
	if input.LimbID != nil {
		limb, err := h.limbRepo.GetByID(*input.LimbID)
		if err == nil {
			limb.Disponibilite = false
			_ = h.limbRepo.Update(limb)
		}
	}

	c.Header("HX-Redirect", "/contracts")
	c.Status(http.StatusCreated)
}

// GetContractDetailsPage affiche le récapitulatif complet d'un contrat spécifique
func (h *ContractHandler) GetContractDetailsPage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID de contrat invalide")
		return
	}

	contract, err := h.contractRepo.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Erreur: le contrat %s n'existe pas!", idStr)
			return
		}
		c.String(http.StatusInternalServerError, "Erreur lors du chargement du contrat %s! Erreur: %v", idStr, err.Error())
		return
	}

	c.HTML(http.StatusOK, "contract_details.html", gin.H{
		"titre":    "Détails du contrat #" + idStr,
		"active":   "contracts",
		"Contract": contract,
		"Version":  config.AppVersion,
	})
}

// DownloadContractPDF génère et envoie le PDF au navigateur
func (h *ContractHandler) DownloadContractPDF(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID de contrat invalide")
		return
	}

	contract, err := h.contractRepo.GetByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.String(http.StatusNotFound, "Erreur 404 : Le contrat n'existe pas.")
			return
		}
		c.String(http.StatusInternalServerError, "Erreur serveur : %v", err)
		return
	}

	// Récupération des réglages via SettingRepository
	settingsList, err := h.settingRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la récupération des réglages : %v", err)
		return
	}

	settingsMap := make(map[string]string)
	for _, s := range settingsList {
		settingsMap[s.Cle] = s.Valeur
		log.Println("contracts.go - boucle de transformation. Clé -> " + s.Cle + " Valeur -> " + s.Valeur)
	}

	filename, err := services.GenerateContractPDF(*contract, settingsMap)
	if err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la création du PDF : %v", err)
		return
	}

	c.FileAttachment(filename, filepath.Base(filename))
}

func isReleasedStatus(statut string) bool {
	return statut == "Terminé" || statut == "Annulé"
}

// UpdateContractStatus modifie l'état d'un contrat via HTMX
func (h *ContractHandler) UpdateContractStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "ID de contrat invalide")
		return
	}

	contract, err := h.contractRepo.GetByID(uint(id))
	if err != nil {
		c.String(http.StatusNotFound, "Contrat introuvable")
		return
	}

	ancienStatut := contract.Statut

	nouveauStatut := c.PostForm("statut")
	nouvelEtatPaiement := c.PostForm("etat_paiement")
	nouveauModePaiement := c.PostForm("mode_paiement")
	nouveauRecuSigne := c.PostForm("recu_signe")

	contract.Statut = nouveauStatut
	contract.EtatPaiement = nouvelEtatPaiement
	contract.ModePaiement = nouveauModePaiement
	if nouveauRecuSigne == "true" {
		contract.RecuSigne = true
	} else {
		contract.RecuSigne = false
	}

	if err := h.contractRepo.Update(contract); err != nil {
		c.String(http.StatusInternalServerError, "Erreur lors de la sauvegarde")
		return
	}

	// LOGIQUE MÉTIER DU STOCK (via repositories injectés)
	if isReleasedStatus(nouveauStatut) && !isReleasedStatus(ancienStatut) {
		if contract.RiserID != nil {
			riser, err := h.riserRepo.GetByID(*contract.RiserID)
			if err == nil {
				riser.Disponibilite = true
				_ = h.riserRepo.Update(riser)
			}
		}
		if contract.LimbID != nil {
			limb, err := h.limbRepo.GetByID(*contract.LimbID)
			if err == nil {
				limb.Disponibilite = true
				_ = h.limbRepo.Update(limb)
			}
		}
	} else if !isReleasedStatus(nouveauStatut) && isReleasedStatus(ancienStatut) {
		if contract.RiserID != nil {
			riser, err := h.riserRepo.GetByID(*contract.RiserID)
			if err == nil {
				riser.Disponibilite = false
				_ = h.riserRepo.Update(riser)
			}
		}
		if contract.LimbID != nil {
			limb, err := h.limbRepo.GetByID(*contract.LimbID)
			if err == nil {
				limb.Disponibilite = false
				_ = h.limbRepo.Update(limb)
			}
		}
	}

	c.Header("HX-Refresh", "true")
	c.Status(http.StatusOK)
}
