package models

import (
	"time"

	"gorm.io/gorm"
)

// User représente un utilisateur ayant accès à l'interface web/API
// pour y réaliser les différentes opérations inhérentes à la location
type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex"`
	Password string
}

// Member représente une personne louant du matériel
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

// MemberRepository définit l'interface de stockage pour les adhérents.
// Tout moteur de stockage (GORM, SQL brut, mock en mémoire pour tests)
// doit implémenter ces méthodes.
type MemberRepository interface {
	GetAll() ([]Member, error)
	GetArchived() ([]Member, error)
	GetByID(id uint) (*Member, error)
	GetByIDWithUnscoped(id uint) (*Member, error)
	Create(member *Member) error
	Update(member *Member) error
	Delete(member *Member) error
	Reactivate(id uint) error
	HardDelete(id uint) error
}

// Limb représente les branches d'un arc
type Limb struct {
	gorm.Model
	NumeroSerie    string `gorm:"uniqueIndex"`
	Marque         string
	Modele         string
	Taille         string
	Puissance      string
	Disponibilite  bool `gorm:"default:true"`
	Commentaire    string
	AnneeAchat     int
	DateInventaire int
	Prix           float64
}

// LimbRepository définit l'interface de stockage pour les branches.
type LimbRepository interface {
	GetAll() ([]Limb, error)
	GetArchived() ([]Limb, error)
	GetByID(id uint) (*Limb, error)
	GetByIDWithUnscoped(id uint) (*Limb, error)
	Create(limb *Limb) error
	Update(limb *Limb) error
	Delete(limb *Limb) error
	Reactivate(id uint) error
	HardDelete(id uint) error
}

// Riser représente la poignée d'un arc
type Riser struct {
	gorm.Model
	NumeroSerie    string `gorm:"uniqueIndex"`
	Marque         string
	Modele         string
	Taille         string
	Lateralite     string
	Couleur        string
	Disponibilite  bool `gorm:"default:true"`
	AnneeAchat     int
	DateInventaire int
	Prix           float64
}

// RiserRepository définit l'interface de stockage pour les poignées.
type RiserRepository interface {
	GetAll() ([]Riser, error)
	GetArchived() ([]Riser, error)
	GetByID(id uint) (*Riser, error)
	GetByIDWithUnscoped(id uint) (*Riser, error)
	Create(riser *Riser) error
	Update(riser *Riser) error
	Delete(riser *Riser) error
	Reactivate(id uint) error
	HardDelete(id uint) error
}

// Setting représente un parametre de notre application
type Setting struct {
	gorm.Model
	Cle    string `gorm:"uniqueIndex"`
	Valeur string
}

// SettingRepository définit l'interface de stockage pour les réglages de l'application.
type SettingRepository interface {
	GetAll() ([]Setting, error)
	GetByKey(key string) (*Setting, error)
	Save(setting *Setting) error
	SaveAll(settings map[string]string) error
}

// Contract représente un contrat liant Member, Limb et riser
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

// ContractRepository définit l'interface de stockage pour les contrats de location.
type ContractRepository interface {
	GetAll() ([]Contract, error)
	GetByID(id uint) (*Contract, error)
	Create(contract *Contract) error
	Update(contract *Contract) error
	Delete(contract *Contract) error
}
