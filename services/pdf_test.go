package services

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/deogracia/toxophilus/models"
)

func TestGenerateContractPDF(t *testing.T) {
	// 1. Préparation d'un contrat fictif complet en mémoire
	contract := models.Contract{}

	// Contournement propre si `ID` est encapsulé dans `gorm.Model`
	contract.ID = 999

	contract.Member = models.Member{
		Nom:          "Robin",
		Prenom:       "Des Bois",
		CodeAdherent: "FR-12345",
	}
	contract.DateDebut = time.Now()
	contract.DateFin = time.Now().AddDate(0, 1, 0) // + 1 mois

	// Ajout des données financières
	contract.MontantLocation = 150.0
	contract.MontantCaution = 300.0
	contract.EtatPaiement = "Payé"
	contract.ModePaiement = "Chèque"

	contract.Riser = models.Riser{
		Marque:      "Hoyt",
		Modele:      "Formula",
		NumeroSerie: "H-999",
	}
	contract.Limb = models.Limb{
		Marque:      "Uukha",
		Modele:      "SX50",
		Taille:      "68",
		Puissance:   "30",
		NumeroSerie: "U-999",
	}
	contract.Accessoires = "Viseur Shibuya, Berger Button"

	// 2. Préparation des "settings" de tests
	mockSettings := map[string]string{
		"pdf_club_name":           "Club de Test",
		"pdf_show_contact_footer": "true",
		"club_address":            "123 Rue de la Cible",
	}

	// 3. Appel de la fonction à tester
	filename, err := GenerateContractPDF(contract, mockSettings)

	// 4. Vérification A : Aucune erreur technique renvoyée par Maroto
	if err != nil {
		t.Fatalf("❌ Erreur inattendue lors de la génération du PDF : %v", err)
	}

	// 5. Vérification B : Le nom du fichier correspond-il à l'ID ?
	expectedFilename := "Contrat-N°" + strconv.FormatInt(int64(contract.ID), 10) + "-" + contract.Member.Nom + "-" + contract.Member.Prenom + "-" + contract.DateDebut.Format("2006-01-02") + ".pdf"
	if filepath.Base(filename) != expectedFilename {
		t.Errorf("❌ Nom de fichier incorrect. Attendu: %s, Obtenu: %s", expectedFilename, filepath.Base(filename))
	}

	// 6. Vérification C : Le fichier a-t-il bien été créé physiquement sur le disque ?
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("❌ Le fichier %s n'a pas été trouvé sur le disque", filename)
	}

	// 7. Nettoyage : On supprime le faux PDF pour garder le dossier propre
	err = os.Remove(filename)
	if err != nil {
		t.Logf("⚠️ Impossible de supprimer le fichier de test %s : %v", filename, err)
	}
}
