package models

import (
	"gorm.io/gorm"
)

// Poignee représente le corps de l'arc (Riser)
type Poignee struct {
	gorm.Model
	Reference  string `gorm:"uniqueIndex;not null" json:"reference"` // Ex: P-001
	Marque     string `json:"marque"`
	Modele     string `json:"modele"`
	Taille     string `json:"taille"`                           // Ex: "25 pouces"
	Lateralite string `json:"lateralite"`                       // "Droitier" ou "Gaucher"
	EnLocation bool   `gorm:"default:false" json:"en_location"` // Permet de savoir si elle est dispo au club
	Remarques  string `json:"remarques"`                        // Pour noter des éclats de peinture ou un filetage abîmé
}

// Branche représente les branches de l'arc (Limbs)
type Branche struct {
	gorm.Model
	Reference  string `gorm:"uniqueIndex;not null" json:"reference"` // Ex: B-001
	Marque     string `json:"marque"`
	Modele     string `json:"modele"`
	Taille     string `json:"taille"`    // Ex: "Courtes", "Moyennes", "Longues" (ou la taille globale ex: 68")
	Puissance  int    `json:"puissance"` // En livres (lbs), ex: 24
	EnLocation bool   `gorm:"default:false" json:"en_location"`
	Remarques  string `json:"remarques"` // Pour noter si elles sont voilées ou abîmées
}
