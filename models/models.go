package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex"`
	Password string
}

type Member struct {
	gorm.Model
	CodeAdherent  string `gorm:"uniqueIndex"`
	Nom           string
	Prenom        string
	DateNaissance time.Time
	Email         string
	Telephone     string
	NumeroRue     string
	Rue           string
	Ville         string
	CodePostal    string
	Contracts     []Contract
}

type Limb struct {
	gorm.Model
	NumeroSerie   string `gorm:"uniqueIndex"`
	Marque        string
	Modele        string
	Taille        string
	Puissance     string
	Disponibilite bool `gorm:"default:true"`
	Commentaire   string
	AnneeAchat    int
	Prix          float64
}

type Riser struct {
	gorm.Model
	NumeroSerie   string `gorm:"uniqueIndex"`
	Marque        string
	Modele        string
	Taille        string
	Lateralite    string
	Couleur       string
	Disponibilite bool `gorm:"default:true"`
	AnneeAchat    int
	Prix          float64
}

type Setting struct {
	gorm.Model
	Cle    string `gorm:"uniqueIndex"`
	Valeur string
}

type Contract struct {
	gorm.Model

	// Le locataire
	MemberID uint   `json:"member_id"`
	Member   Member `json:"member"`

	// Le Statut global du contrat (NOUVEAU)
	Statut string `gorm:"default:'Actif'" json:"statut"` // Ex: "Actif", "Expiré", "Terminé", "Annulé"

	// Périodes
	DateDebut time.Time `json:"date_debut"`
	DateFin   time.Time `json:"date_fin"`

	// Le matériel (Pointeurs pour être facultatifs / NULL en base)
	// on peut louer soit branches et poignées, soit l'un ou soit l'autre.
	RiserID *uint `json:"riser_id"`
	Riser   Riser `json:"riser"`
	LimbID  *uint `json:"limb_id"`
	Limb    Limb  `json:"limb"`

	// Équipements additionnels et remarques pour le PDF
	Accessoires string `json:"accessoires"` // Ex: "Viseur, Berger Button"
	Commentaire string `json:"commentaire"` // Remarques spécifiques affichées sur le contrat

	// Aspect financier et paiement
	MontantLocation   float64 `json:"montant_location"`
	EtatPaiement      string  `json:"etat_paiement"`       // Ex: "En attente", "Payé"
	ModePaiement      string  `json:"mode_paiement"`       // Ex: "CB", "Chèque", "Espèces", "Autre"
	ModePaiementAutre string  `json:"mode_paiement_autre"` // Précision si mode "Autre"
	MontantCaution    float64 `json:"montant_caution"`

	// Suivi administratif
	RecuSigne       bool   `gorm:"default:false" json:"recu_signe"`
	CheminPDFGenere string `json:"chemin_pdf_genere"` // Contrat vierge issu de l'application
	CheminPDFSigne  string `json:"chemin_pdf_signe"`  // Contrat signé retourné (scan/photo)
}
