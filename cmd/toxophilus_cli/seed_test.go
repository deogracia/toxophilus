package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSeedData(t *testing.T) {
	// 1. Initialisation d'une base SQLite en mémoire pour le test
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Impossible d'ouvrir la DB de test : %v", err)
	}

	// 2. Migration des modèles requis
	err = db.AutoMigrate(&models.Member{}, &models.Riser{}, &models.Limb{}, &models.Contract{}, &models.Setting{})
	if err != nil {
		t.Fatalf("Impossible de migrer les modèles : %v", err)
	}

	// On associe la base globale temporairement pour le test
	database.DB = db

	// 3. Désérialisation du JSON de seed
	var data SeedData
	if err := json.Unmarshal(seedJSON, &data); err != nil {
		t.Fatalf("Erreur lors du parsing du JSON de seed : %v", err)
	}

	if len(data.Members) == 0 {
		t.Fatal("Le JSON de seed ne devrait pas contenir 0 membres")
	}

	// 4. Lancement du seeding
	seedData()

	// 5. Assertions de validation
	var memberCount int64
	db.Model(&models.Member{}).Count(&memberCount)
	if memberCount != int64(len(data.Members)) {
		t.Errorf("Attendu %d membres en base, obtenu %d", len(data.Members), memberCount)
	}

	var riserCount int64
	db.Model(&models.Riser{}).Count(&riserCount)
	if riserCount != int64(len(data.Risers)) {
		t.Errorf("Attendu %d poignées en base, obtenu %d", len(data.Risers), riserCount)
	}

	var limbCount int64
	db.Model(&models.Limb{}).Count(&limbCount)
	if limbCount != int64(len(data.Limbs)) {
		t.Errorf("Attendu %d branches en base, obtenu %d", len(data.Limbs), limbCount)
	}

	var contractCount int64
	db.Model(&models.Contract{}).Count(&contractCount)
	if contractCount != int64(len(data.Contracts)) {
		t.Errorf("Attendu %d contrats en base, obtenu %d", len(data.Contracts), contractCount)
	}

	// Vérification de la mise à jour des stocks pour le matériel loué (Contrat Actif)
	var activeRiser models.Riser
	db.Where("numero_serie = ?", "P-HOYT-2026").First(&activeRiser)
	if activeRiser.Disponibilite {
		t.Error("La poignée Hoyt (liée à un contrat Actif) devrait être marquée comme non disponible")
	}

	var activeLimb models.Limb
	db.Where("numero_serie = ?", "B-UUKHA-2026").First(&activeLimb)
	if activeLimb.Disponibilite {
		t.Error("Les branches Uukha (liées à un contrat Actif) devraient être marquées comme non disponibles")
	}

	// Vérification que le matériel restitué (Contrat Terminé) est disponible
	var finishedRiser models.Riser
	db.Where("numero_serie = ?", "P-WNS-2026").First(&finishedRiser)
	if !finishedRiser.Disponibilite {
		t.Error("La poignée WNS (liée à un contrat Terminé) devrait être marquée comme disponible")
	}

	// Vérification de l'injection des réglages de démonstration (Settings)
	var setting models.Setting
	db.Where("cle = ?", "pdf_club_name").First(&setting)
	if setting.Valeur != "Club de Tir à l'Arc - Toxophilus" {
		t.Errorf("Attendu 'Club de Tir à l'Arc - Toxophilus' pour pdf_club_name, obtenu %s", setting.Valeur)
	}

	var settingSubtitle models.Setting
	db.Where("cle = ?", "pdf_club_subtitle").First(&settingSubtitle)
	if settingSubtitle.Valeur != "La flèche de Sénart" {
		t.Errorf("Attendu 'La flèche de Sénart' pour pdf_club_subtitle, obtenu %s", settingSubtitle.Valeur)
	}

	var settingClause models.Setting
	db.Where("cle = ?", "pdf_clause_mise_disposition").First(&settingClause)
	if !strings.Contains(settingClause.Valeur, "Le club, dont les données figurent en bas de ce contrat, met à disposition de l'adhérent le matériel d'archerie désigné ci-dessus") {
		t.Errorf("La clause de mise à disposition est absente ou incorrecte, obtenu %s", settingClause.Valeur)
	}
}
