package auth

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSelfSignedCert(t *testing.T) {
	// Création d'un dossier temporaire pour les tests de certificats
	tmpDir, err := os.MkdirTemp("", "toxo_tls_test_*")
	if err != nil {
		t.Fatalf("Impossible de créer le dossier temporaire : %v", err)
	}
	defer os.RemoveAll(tmpDir)

	certPath := filepath.Join(tmpDir, "test_localhost.crt")
	keyPath := filepath.Join(tmpDir, "test_localhost.key")

	// 1. Première génération
	err = EnsureSelfSignedCert(certPath, keyPath)
	if err != nil {
		t.Fatalf("La génération du certificat a échoué : %v", err)
	}

	// Vérification de l'existence des fichiers
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Error("Le fichier de certificat n'a pas été créé")
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Error("Le fichier de clé privée n'a pas été créé")
	}

	// 2. Lecture et validation du contenu du certificat
	certBytes, err := os.ReadFile(certPath)
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

	// 3. Appel secondaire (ne doit pas régénérer car déjà existant et valide)
	err = EnsureSelfSignedCert(certPath, keyPath)
	if err != nil {
		t.Fatalf("Le second appel de vérification a échoué : %v", err)
	}
}
