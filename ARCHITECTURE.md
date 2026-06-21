# 🏹 Architecture et Décisions Techniques - Gestion Archerie

# 1. Vue d'ensemble

L'application web a pour but de gérer le parc matériel (poignées, branches) et les contrats de location d'un club de tir à l'arc à ses licenciés. Elle permet la génération de contrats PDF, l'envoi d'emails et la gestion d'une tarification dynamique.

# 2. Choix Technologiques

    * Langage : Go (Golang). Choisi pour sa légèreté, ses performances et sa capacité à compiler l'application en un seul fichier binaire exécutable, facilitant grandement le déploiement.

    * Framework Web : Gin. Utilisé pour gérer les routes HTTP et l'API de manière performante.

    * ORM (Base de données) : GORM. Choisi pour sa facilité de migration automatique des schémas et son abstraction du moteur de base de données.

    * Base de données : Architecture hybride.

       * SQLite par défaut (stockage dans un fichier local club_archerie.db, idéal pour le club).

       * MySQL / PostgreSQL supportés via modification de la configuration pour une éventuelle évolution vers un serveur dédié.

    * Gestion de la Configuration : Architecture à deux niveaux.

       * Infrastructure (Viper + TOML) : Pour les paramètres de démarrage (port, identifiants BDD).

       * Métier (GORM en BDD) : Pour les règles de gestion (prix des arcs, durées par défaut), modifiables à chaud sans redémarrer le serveur.

# 3. Schéma de Données (Modèles)

```go
// models/models.go
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

// Limb représente les branches d'un arc
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

// Riser représente la poignée d'un arc
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

// Setting représente un parametre de notre application
type Setting struct {
	gorm.Model
	Cle    string `gorm:"uniqueIndex"`
	Valeur string
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
```

# 4. Implémentation Core

## 4.1 Fichier de Configuration Infrastructure (`config.toml`)

À placer à la racine du projet. Surchargable par des variables d'environnement.

## 4.2 Chargement de la configuration (`config/load.go`)

## 4.3 Connexion Dynamique à la Base de Données (`database/connect.go`)

## 4.4 Gestion des Paramètres Métier (`services/settings.go`)

## 4.5 Point d'entrée de l'application (`main.go`)