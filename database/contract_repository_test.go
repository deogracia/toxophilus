package database

import (
	"testing"
	"time"

	"github.com/deogracia/toxophilus/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupContractRepoTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Impossible d'ouvrir la DB de test : %v", err)
	}

	err = db.AutoMigrate(&models.Member{}, &models.Riser{}, &models.Limb{}, &models.Contract{}, &models.Setting{})
	if err != nil {
		t.Fatalf("Impossible de migrer les tables : %v", err)
	}

	return db
}

func TestGormContractRepository(t *testing.T) {
	db := setupContractRepoTestDB(t)
	repo := NewGormContractRepository(db)

	member := models.Member{CodeAdherent: "TEST-01", Nom: "Doe", Prenom: "John"}
	db.Create(&member)

	t.Run("Create & GetByID", func(t *testing.T) {
		contract := &models.Contract{
			MemberID:        member.ID,
			Statut:          "Actif",
			DateDebut:       time.Now(),
			DateFin:         time.Now().Add(24 * time.Hour),
			MontantLocation: 100,
		}

		err := repo.Create(contract)
		if err != nil {
			t.Fatalf("Erreur de création du contrat : %v", err)
		}

		fetched, err := repo.GetByID(contract.ID)
		if err != nil {
			t.Fatalf("Erreur de récupération : %v", err)
		}

		if fetched.Member.Nom != "Doe" {
			t.Errorf("Attendu Doe, obtenu %s", fetched.Member.Nom)
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		contracts, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Erreur GetAll : %v", err)
		}

		if len(contracts) != 1 {
			t.Errorf("Attendu 1 contrat, obtenu %d", len(contracts))
		}
	})

	t.Run("Update", func(t *testing.T) {
		contracts, _ := repo.GetAll()
		contract := &contracts[0]

		contract.Statut = "Terminé"
		err := repo.Update(contract)
		if err != nil {
			t.Fatalf("Erreur de mise à jour : %v", err)
		}

		fetched, _ := repo.GetByID(contract.ID)
		if fetched.Statut != "Terminé" {
			t.Errorf("Attendu Terminé, obtenu %s", fetched.Statut)
		}
	})
}

func TestGormSettingRepository(t *testing.T) {
	db := setupContractRepoTestDB(t)
	repo := NewGormSettingRepository(db)

	t.Run("SaveAll & GetAll & GetByKey", func(t *testing.T) {
		settings := map[string]string{
			"pdf_club_name":     "Archerie Club",
			"pdf_club_subtitle": "Toxophilus",
		}

		err := repo.SaveAll(settings)
		if err != nil {
			t.Fatalf("Erreur SaveAll : %v", err)
		}

		all, err := repo.GetAll()
		if err != nil {
			t.Fatalf("Erreur GetAll : %v", err)
		}

		if len(all) != 2 {
			t.Errorf("Attendu 2 réglages, obtenu %d", len(all))
		}

		fetched, err := repo.GetByKey("pdf_club_name")
		if err != nil {
			t.Fatalf("Erreur GetByKey : %v", err)
		}

		if fetched.Valeur != "Archerie Club" {
			t.Errorf("Attendu 'Archerie Club', obtenu %s", fetched.Valeur)
		}
	})
}
