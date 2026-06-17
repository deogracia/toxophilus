package database

import (
	"testing"

	"github.com/deogracia/toxophilus/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupEquipementRepoTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Impossible d'ouvrir la DB de test : %v", err)
	}

	err = db.AutoMigrate(&models.Riser{}, &models.Limb{})
	if err != nil {
		t.Fatalf("Impossible de migrer les tables d'équipement : %v", err)
	}

	return db
}

func TestGormRiserRepository(t *testing.T) {
	db := setupEquipementRepoTestDB(t)
	repo := NewGormRiserRepository(db)

	t.Run("Create & GetByID", func(t *testing.T) {
		riser := &models.Riser{
			NumeroSerie: "R-101",
			Marque:      "Hoyt",
			Modele:      "Formula",
			Taille:      "25 pouces",
			Lateralite:  "Droitier",
		}

		err := repo.Create(riser)
		if err != nil {
			t.Fatalf("Erreur de création : %v", err)
		}

		fetched, err := repo.GetByID(riser.ID)
		if err != nil {
			t.Fatalf("Erreur de récupération : %v", err)
		}

		if fetched.NumeroSerie != "R-101" {
			t.Errorf("Attendu R-101, obtenu %s", fetched.NumeroSerie)
		}
	})

	t.Run("Soft Delete & Reactivate", func(t *testing.T) {
		riser := &models.Riser{NumeroSerie: "R-102", Marque: "WNS"}
		_ = repo.Create(riser)

		_ = repo.Delete(riser)

		_, err := repo.GetByID(riser.ID)
		if err == nil {
			t.Error("Ne devrait pas être trouvable après soft delete")
		}

		_ = repo.Reactivate(riser.ID)

		fetched, err := repo.GetByID(riser.ID)
		if err != nil {
			t.Fatalf("Erreur après réactivation : %v", err)
		}
		if fetched.NumeroSerie != "R-102" {
			t.Errorf("Attendu R-102, obtenu %s", fetched.NumeroSerie)
		}
	})
}

func TestGormLimbRepository(t *testing.T) {
	db := setupEquipementRepoTestDB(t)
	repo := NewGormLimbRepository(db)

	t.Run("Create & GetByID", func(t *testing.T) {
		limb := &models.Limb{
			NumeroSerie: "L-101",
			Marque:      "Uukha",
			Taille:      "68 pouces",
			Puissance:   "28#",
		}

		err := repo.Create(limb)
		if err != nil {
			t.Fatalf("Erreur de création : %v", err)
		}

		fetched, err := repo.GetByID(limb.ID)
		if err != nil {
			t.Fatalf("Erreur de récupération : %v", err)
		}

		if fetched.NumeroSerie != "L-101" {
			t.Errorf("Attendu L-101, obtenu %s", fetched.NumeroSerie)
		}
	})
}
