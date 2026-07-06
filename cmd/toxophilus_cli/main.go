package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/auth"
	"github.com/deogracia/toxophilus/models"
)

//go:embed seed.json
var seedJSON []byte

type SeedMember struct {
	CodeAdherent          string    `json:"code_adherent"`
	Nom                   string    `json:"nom"`
	Prenom                string    `json:"prenom"`
	DateNaissance         time.Time `json:"date_naissance"`
	Email                 string    `json:"email"`
	Telephone             string    `json:"telephone"`
	NumeroRue             string    `json:"numero_rue"`
	Rue                   string    `json:"rue"`
	Ville                 string    `json:"ville"`
	CodePostal            string    `json:"code_postal"`
	ParentNom             string    `json:"parent_nom"`
	ParentPrenom          string    `json:"parent_prenom"`
	ParentTelephone       string    `json:"parent_telephone"`
	ParentEmail           string    `json:"parent_email"`
	ParentRelation        string    `json:"parent_relation"`
	EstEmancipe           bool      `json:"est_emancipe"`
	ReferenceEmancipation string    `json:"reference_emancipation"`
}

type SeedRiser struct {
	NumeroSerie    string  `json:"numero_serie"`
	Marque         string  `json:"marque"`
	Modele         string  `json:"modele"`
	Taille         string  `json:"taille"`
	Lateralite     string  `json:"lateralite"`
	Couleur        string  `json:"couleur"`
	Disponibilite  bool    `json:"disponibilite"`
	AnneeAchat     int     `json:"annee_achat"`
	DateInventaire int     `json:"date_inventaire"`
	Prix           float64 `json:"prix"`
}

type SeedLimb struct {
	NumeroSerie    string  `json:"numero_serie"`
	Marque         string  `json:"marque"`
	Modele         string  `json:"modele"`
	Taille         string  `json:"taille"`
	Puissance      string  `json:"puissance"`
	Disponibilite  bool    `json:"disponibilite"`
	Commentaire    string  `json:"commentaire"`
	AnneeAchat     int     `json:"annee_achat"`
	DateInventaire int     `json:"date_inventaire"`
	Prix           float64 `json:"prix"`
}

type SeedData struct {
	Members   []SeedMember      `json:"members"`
	Risers    []SeedRiser       `json:"risers"`
	Limbs     []SeedLimb        `json:"limbs"`
	Settings  map[string]string `json:"settings"`
	Contracts []struct {
		MemberCode       string  `json:"member_code"`
		RiserSerial      string  `json:"riser_serial"`
		LimbSerial       string  `json:"limb_serial"`
		Statut           string  `json:"statut"`
		OffsetDebutJours int     `json:"offset_debut_jours"`
		OffsetFinMois    int     `json:"offset_fin_mois"`
		OffsetFinJours   int     `json:"offset_fin_jours"`
		Accessoires      string  `json:"accessoires"`
		Commentaire      string  `json:"commentaire"`
		MontantLocation  float64 `json:"montant_location"`
		MontantCaution   float64 `json:"montant_caution"`
		EtatPaiement     string  `json:"etat_paiement"`
		ModePaiement     string  `json:"mode_paiement"`
		RecuSigne        bool    `json:"recu_signe"`
	} `json:"contracts"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	subcommand := os.Args[1]

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("❌ Impossible de charger la configuration.\n Erreur: %v", err)
	}

	database.Connect()

	switch subcommand {
	case "create-admin":
		if len(os.Args) < 4 {
			fmt.Println("❌ Usage: toxophilus_cli create-admin <email> <password>")
			return
		}
		createAdmin(os.Args[2], os.Args[3])

	case "seed":
		seedData()

	case "generate-certs":
		if err := auth.EnsureSelfSignedCert(); err != nil {
			log.Fatalf("❌ Erreur lors de la génération des certificats : %v", err)
		}

	default:
		fmt.Printf("❌ Sous-commande %q inconnue.\n", subcommand)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("🏹 Outil CLI Toxophilus")
	fmt.Println("Usage:")
	fmt.Println("  toxophilus_cli create-admin <email> <password>  - Crée un administrateur")
	fmt.Println("  toxophilus_cli seed                             - Injecte des données factices de démonstration")
	fmt.Println("  toxophilus_cli generate-certs [cert] [key]      - Génère des certificats TLS auto-signés pour localhost")
}

func createAdmin(email, password string) {
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("❌ Erreur de hachage du mot de passe : %v", err)
	}

	user := models.User{Email: email, Password: hashedPassword}

	// On vérifie si l'utilisateur existe déjà
	var count int64
	database.DB.Model(&models.User{}).Where("email = ?", email).Count(&count)
	if count > 0 {
		fmt.Printf("⚠️ L'administrateur %s existe déjà.\n", email)
		return
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Fatalf("❌ Erreur de création de l'administrateur : %v", err)
	}

	fmt.Printf("✅ Administrateur %s créé avec succès !\n", email)
}

func seedData() {
	fmt.Println("🌱 Chargement et injection des données de démonstration...")

	var data SeedData
	if err := json.Unmarshal(seedJSON, &data); err != nil {
		log.Fatalf("❌ Impossible de parser le fichier de seed : %v", err)
	}

	// 1. MEMBRES
	memberMap := make(map[string]models.Member)
	for _, sm := range data.Members {
		var existing models.Member
		err := database.DB.Where("code_adherent = ?", sm.CodeAdherent).First(&existing).Error
		if err != nil {
			m := models.Member{
				CodeAdherent:          sm.CodeAdherent,
				Nom:                   sm.Nom,
				Prenom:                sm.Prenom,
				DateNaissance:         sm.DateNaissance,
				Email:                 sm.Email,
				Telephone:             sm.Telephone,
				NumeroRue:             sm.NumeroRue,
				Rue:                   sm.Rue,
				Ville:                 sm.Ville,
				CodePostal:            sm.CodePostal,
				ParentNom:             sm.ParentNom,
				ParentPrenom:          sm.ParentPrenom,
				ParentTelephone:       sm.ParentTelephone,
				ParentEmail:           sm.ParentEmail,
				ParentRelation:        sm.ParentRelation,
				EstEmancipe:           sm.EstEmancipe,
				ReferenceEmancipation: sm.ReferenceEmancipation,
			}
			if err := database.DB.Create(&m).Error; err != nil {
				log.Fatalf("❌ Impossible de créer le membre %s : %v", sm.CodeAdherent, err)
			}
			memberMap[sm.CodeAdherent] = m
			fmt.Printf("   -> Membre %s %s (%s) créé.\n", m.Prenom, m.Nom, m.CodeAdherent)
		} else {
			memberMap[sm.CodeAdherent] = existing
			fmt.Printf("   -> Membre %s déjà existant.\n", sm.CodeAdherent)
		}
	}

	// 2. POIGNÉES (RISERS)
	riserMap := make(map[string]models.Riser)
	for _, sr := range data.Risers {
		var existing models.Riser
		err := database.DB.Where("numero_serie = ?", sr.NumeroSerie).First(&existing).Error
		if err != nil {
			r := models.Riser{
				NumeroSerie:    sr.NumeroSerie,
				Marque:         sr.Marque,
				Modele:         sr.Modele,
				Taille:         sr.Taille,
				Lateralite:     sr.Lateralite,
				Couleur:        sr.Couleur,
				Disponibilite:  sr.Disponibilite,
				AnneeAchat:     sr.AnneeAchat,
				DateInventaire: sr.DateInventaire,
				Prix:           sr.Prix,
			}
			if err := database.DB.Create(&r).Error; err != nil {
				log.Fatalf("❌ Impossible de créer la poignée %s : %v", sr.NumeroSerie, err)
			}
			riserMap[sr.NumeroSerie] = r
			fmt.Printf("   -> Poignée %s %s (%s) créée.\n", r.Marque, r.Modele, r.NumeroSerie)
		} else {
			riserMap[sr.NumeroSerie] = existing
			fmt.Printf("   -> Poignée %s déjà existante.\n", sr.NumeroSerie)
		}
	}

	// 3. BRANCHES (LIMBS)
	limbMap := make(map[string]models.Limb)
	for _, sl := range data.Limbs {
		var existing models.Limb
		err := database.DB.Where("numero_serie = ?", sl.NumeroSerie).First(&existing).Error
		if err != nil {
			l := models.Limb{
				NumeroSerie:    sl.NumeroSerie,
				Marque:         sl.Marque,
				Modele:         sl.Modele,
				Taille:         sl.Taille,
				Puissance:      sl.Puissance,
				Disponibilite:  sl.Disponibilite,
				Commentaire:    sl.Commentaire,
				AnneeAchat:     sl.AnneeAchat,
				DateInventaire: sl.DateInventaire,
				Prix:           sl.Prix,
			}
			if err := database.DB.Create(&l).Error; err != nil {
				log.Fatalf("❌ Impossible de créer les branches %s : %v", sl.NumeroSerie, err)
			}
			limbMap[sl.NumeroSerie] = l
			fmt.Printf("   -> Branches %s %s (%s) créées.\n", l.Marque, l.Modele, l.NumeroSerie)
		} else {
			limbMap[sl.NumeroSerie] = existing
			fmt.Printf("   -> Branches %s déjà existantes.\n", sl.NumeroSerie)
		}
	}

	// 4. CONTRATS
	for _, c := range data.Contracts {
		member, ok := memberMap[c.MemberCode]
		if !ok {
			fmt.Printf("⚠️ Membre %s introuvable pour le contrat. Ignoré.\n", c.MemberCode)
			continue
		}

		// On regarde si un contrat actif ou similaire existe déjà pour éviter les doublons de seeding
		var existingContractCount int64
		database.DB.Model(&models.Contract{}).Where("member_id = ? AND statut = ?", member.ID, c.Statut).Count(&existingContractCount)
		if existingContractCount > 0 {
			fmt.Printf("   -> Contrat %s pour le membre %s déjà existant.\n", c.Statut, c.MemberCode)
			continue
		}

		var riserID *uint
		if c.RiserSerial != "" {
			if riser, ok := riserMap[c.RiserSerial]; ok {
				riserID = &riser.ID
			}
		}

		var limbID *uint
		if c.LimbSerial != "" {
			if limb, ok := limbMap[c.LimbSerial]; ok {
				limbID = &limb.ID
			}
		}

		// Calcul des dates dynamiques
		debut := time.Now().AddDate(0, 0, c.OffsetDebutJours)
		fin := time.Now().AddDate(0, c.OffsetFinMois, c.OffsetFinJours)

		contract := models.Contract{
			MemberID:        member.ID,
			Statut:          c.Statut,
			DateDebut:       debut,
			DateFin:         fin,
			RiserID:         riserID,
			LimbID:          limbID,
			Accessoires:     c.Accessoires,
			Commentaire:     c.Commentaire,
			MontantLocation: c.MontantLocation,
			MontantCaution:  c.MontantCaution,
			EtatPaiement:    c.EtatPaiement,
			ModePaiement:    c.ModePaiement,
			RecuSigne:       c.RecuSigne,
		}

		if err := database.DB.Omit("Riser", "Limb").Create(&contract).Error; err != nil {
			log.Fatalf("❌ Impossible de créer le contrat : %v", err)
		}

		// Mise à jour de la disponibilité en stock si le contrat est actif
		disponibiliteMatériel := true
		if c.Statut == "Actif" {
			disponibiliteMatériel = false
		}

		if riserID != nil {
			database.DB.Model(&models.Riser{}).Where("id = ?", *riserID).Update("disponibilite", disponibiliteMatériel)
		}
		if limbID != nil {
			database.DB.Model(&models.Limb{}).Where("id = ?", *limbID).Update("disponibilite", disponibiliteMatériel)
		}

		fmt.Printf("   -> Contrat %s de %s %s enregistré (Dates : %s au %s).\n",
			c.Statut, member.Prenom, member.Nom, debut.Format("02/01/2006"), fin.Format("02/01/2006"))
	}

	// 5. PARAMÈTRES (SETTINGS)
	for k, v := range data.Settings {
		var existing models.Setting
		err := database.DB.Where("cle = ?", k).First(&existing).Error
		if err != nil {
			setting := models.Setting{Cle: k, Valeur: v}
			if err := database.DB.Create(&setting).Error; err == nil {
				fmt.Printf("   -> Réglage %s créé.\n", k)
			}
		} else {
			existing.Valeur = v
			database.DB.Save(&existing)
			fmt.Printf("   -> Réglage %s mis à jour.\n", k)
		}
	}

	fmt.Println("✅ Injection de données factices terminée avec succès !")
}
