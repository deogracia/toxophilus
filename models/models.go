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
	ContractID    *uint
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
	ContractID    *uint
}

type Contract struct {
	gorm.Model
	MemberID          uint
	Member            Member
	TypeContrat       string
	MontantLocation   float64
	EtatPaiement      string
	ModePaiement      string
	ModePaiementAutre string
	MontantCaution    float64
	DateDebut         time.Time
	DateFin           time.Time
	RecuSigne         bool    `gorm:"default:false"`
	Risers            []Riser `gorm:"foreignKey:ContractID"`
	Limbs             []Limb  `gorm:"foreignKey:ContractID"`
}

type Setting struct {
	gorm.Model
	Cle    string `gorm:"uniqueIndex"`
	Valeur string
}
