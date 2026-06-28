package auth

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

func TestEnsureSelfSignedCert(t *testing.T) {
	// 1. Sauvegarde des certificats existants éventuels de l'utilisateur
	backupCert, errCert := os.ReadFile(CertPath)
	backupKey, errKey := os.ReadFile(KeyPath)

	// On nettoie les fichiers temporairement pour forcer la génération
	_ = os.Remove(CertPath)
	_ = os.Remove(KeyPath)

	// Au nettoyage de fin de test, on s'assure de restaurer la situation de départ ou de supprimer nos fichiers de test
	defer func() {
		_ = os.Remove(CertPath)
		_ = os.Remove(KeyPath)
		if errCert == nil {
			_ = os.WriteFile(CertPath, backupCert, 0600)
		}
		if errKey == nil {
			_ = os.WriteFile(KeyPath, backupKey, 0600)
		}
	}()

	// 2. Première génération
	err := EnsureSelfSignedCert()
	if err != nil {
		t.Fatalf("La génération du certificat a échoué : %v", err)
	}

	// Vérification de l'existence des fichiers
	if _, err := os.Stat(CertPath); os.IsNotExist(err) {
		t.Error("Le fichier de certificat n'a pas été créé")
	}
	if _, err := os.Stat(KeyPath); os.IsNotExist(err) {
		t.Error("Le fichier de clé privée n'a pas été créé")
	}

	// 3. Lecture et validation du contenu du certificat
	certBytes, err := os.ReadFile(CertPath)
	if err != nil {
		t.Fatalf("Impossible de lire le certificat généré : %v", err)
	}

	block, _ := pem.Decode(certBytes)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatal("Le fichier généré n'est pas un certificat PEM valide")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("Impossible de parser le certificat x509 : %v", err)
	}

	if cert.Subject.CommonName != "localhost" {
		t.Errorf("CommonName attendu: localhost, obtenu: %s", cert.Subject.CommonName)
	}

	// 4. Appel secondaire (ne doit pas régénérer car déjà existant et valide)
	err = EnsureSelfSignedCert()
	if err != nil {
		t.Fatalf("Le second appel de vérification a échoué : %v", err)
	}
}
