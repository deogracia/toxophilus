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

```Go
// models/models.go
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
	RecuSigne         bool `gorm:"default:false"` 
	Risers            []Riser `gorm:"foreignKey:ContractID"`
	Limbs             []Limb  `gorm:"foreignKey:ContractID"`
}

type Setting struct {
	gorm.Model
	Cle    string `gorm:"uniqueIndex"` 
	Valeur string
}
```

# 4. Implémentation Core

## 4.1 Fichier de Configuration Infrastructure (`config.toml`)

À placer à la racine du projet. Surchargable par des variables d'environnement.

## 4.2 Chargement de la configuration (`config/load.go`)

## 4.3 Connexion Dynamique à la Base de Données (`database/connect.go`)

## 4.4 Gestion des Paramètres Métier (`services/settings.go`)

## 4.5 Point d'entrée de l'application (`main.go`)