package database

import (
	"testing"
	"time"

	"github.com/deogracia/toxophilus/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupRepoTestDB initialise une base SQLite en mémoire pour tester le repository
func setupRepoTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Impossible d'ouvrir la DB de test : %v", err)
	}

	err = db.AutoMigrate(&models.Member{})
	if err != nil {
		t.Fatalf("Impossible de migrer la table Member : %v", err)
	}

	return db
}

func TestGormMemberRepository(t *testing.T) {
	db := setupRepoTestDB(t)
	repo := NewGormMemberRepository(db)

	t.Run("Create & GetByID", func(t *testing.T) {
		member := &models.Member{
			CodeAdherent:  "M-101",
			Nom:           "Lannister",
			Prenom:        "Jaime",
			DateNaissance: time.Now(),
			Email:         "jaime@kingslanding.com",
		}

		err := repo.Create(member)
		if err != nil {
			t.Fatalf("Erreur lors de la création : %v", err)
		}

		if member.ID == 0 {
			t.Error("L'ID du membre créé ne devrait pas être 0")
		}

		fetched, err := repo.GetByID(member.ID)
		if err != nil {
			t.Fatalf("Erreur lors de la récupération : %v", err)
		}

		if fetched.Nom != "Lannister" || fetched.Prenom != "Jaime" {
			t.Errorf("Propriétés incorrectes : attendu Lannister Jaime, obtenu %s %s", fetched.Nom, fetched.Prenom)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		// On réinitialise pour ce sous-test
		db := setupRepoTestDB(t)
		repo := NewGormMemberRepository(db)

		_ = repo.Create(&models.Member{CodeAdherent: "M-1", Nom: "A"})
		_ = repo.Create(&models.Member{CodeAdherent: "M-2", Nom: "B"})

		members, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Erreur lors de la récupération : %v", err)
		}

		if len(members) != 2 {
			t.Errorf("Attendu 2 membres, obtenu %d", len(members))
		}
	})

	t.Run("Update", func(t *testing.T) {
		member := &models.Member{
			CodeAdherent: "M-102",
			Nom:          "Stark",
			Prenom:       "Sansa",
		}
		_ = repo.Create(member)

		member.Prenom = "Alayne"
		err := repo.Update(member)
		if err != nil {
			t.Fatalf("Erreur lors de la mise à jour : %v", err)
		}

		fetched, _ := repo.GetByID(member.ID)
		if fetched.Prenom != "Alayne" {
			t.Errorf("Attendu Alayne, obtenu %s", fetched.Prenom)
		}
	})

	t.Run("Delete (Soft Delete) & Reactivate", func(t *testing.T) {
		member := &models.Member{
			CodeAdherent: "M-103",
			Nom:          "Greyjoy",
			Prenom:       "Theon",
		}
		_ = repo.Create(member)

		// Soft Delete
		err := repo.Delete(member)
		if err != nil {
			t.Fatalf("Erreur lors de la suppression : %v", err)
		}

		// Vérification qu'on ne le trouve plus de manière normale
		_, err = repo.GetByID(member.ID)
		if err == nil {
			t.Error("On ne devrait pas pouvoir récupérer un membre soft-deleted")
		}

		// Vérification qu'il apparaît dans les archives
		archived, err := repo.GetArchived()
		if err != nil {
			t.Fatalf("Erreur lors de la récupération des archives : %v", err)
		}
		if len(archived) != 1 || archived[0].ID != member.ID {
			t.Error("Le membre soft-deleted devrait figurer dans les archives")
		}

		// Réactivation
		err = repo.Reactivate(member.ID)
		if err != nil {
			t.Fatalf("Erreur lors de la réactivation : %v", err)
		}

		// Vérification qu'on peut à nouveau le récupérer normalement
		reactivated, err := repo.GetByID(member.ID)
		if err != nil {
			t.Fatalf("Erreur après réactivation : %v", err)
		}
		if reactivated.Prenom != "Theon" {
			t.Errorf("Nom incorrect après réactivation : %s", reactivated.Prenom)
		}
	})

	t.Run("HardDelete", func(t *testing.T) {
		member := &models.Member{
			CodeAdherent: "M-104",
			Nom:          "Clegane",
			Prenom:       "Sandor",
		}
		_ = repo.Create(member)

		err := repo.HardDelete(member.ID)
		if err != nil {
			t.Fatalf("Erreur lors de la suppression définitive : %v", err)
		}

		// Vérification qu'il n'existe absolument plus (même en Unscoped)
		_, err = repo.GetByIDWithUnscoped(member.ID)
		if err == nil {
			t.Error("Le membre supprimé définitivement ne devrait plus exister du tout")
		}
	})
}
