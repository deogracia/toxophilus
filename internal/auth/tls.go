package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// Constantes strictes de chemins pour la génération de certificats TLS en développement local.
// L'utilisation de constantes fixes évite toute vulnérabilité d'inclusion ou d'injection de chemin (gosec G304).
const (
	CertPath = "localhost.crt"
	KeyPath  = "localhost.key"
)

// EnsureSelfSignedCert s'assure qu'un certificat auto-signé valide existe pour localhost à la racine.
// S'il n'existe pas, ou s'il expire dans moins de 30 jours, il est généré/renouvelé automatiquement.
func EnsureSelfSignedCert() error {
	regenerate := false

	// 1. Vérifier si les fichiers existent déjà
	if _, err := os.Stat(CertPath); err != nil {
		regenerate = true
	} else if _, err := os.Stat(KeyPath); err != nil {
		regenerate = true
	} else {
		// Les fichiers existent, vérifions leur expiration pour un renouvellement automatique
		certBytes, err := os.ReadFile(CertPath)
		if err != nil {
			regenerate = true
		} else {
			block, _ := pem.Decode(certBytes)
			if block == nil || block.Type != "CERTIFICATE" {
				regenerate = true
			} else {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					regenerate = true
				} else {
					// Si le certificat expire dans moins de 30 jours, on planifie le renouvellement automatique
					if time.Now().Add(30 * 24 * time.Hour).After(cert.NotAfter) {
						regenerate = true
					}
				}
			}
		}
	}

	if !regenerate {
		return nil
	}

	fmt.Printf("Generating self-signed TLS certificate for localhost (valid for 1 year)...\n")

	// S'assurer que le dossier parent existe (avec des permissions restrictives 0700 pour corriger G301)
	if dir := filepath.Dir(CertPath); dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	if dir := filepath.Dir(KeyPath); dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	// 2. Génération de la clé privée RSA de 2048 bits
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 3. Configuration du certificat pour localhost / 127.0.0.1 (validité 1 an)
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valide pendant 1 an

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Toxophilus Local Dev"},
			CommonName:   "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.IPv6loopback},
		DNSNames:              []string{"localhost"},
	}

	// 4. Création du certificat auto-signé
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// 5. Écriture du certificat au format PEM
	certOut, err := os.Create(CertPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	// 6. Écriture de la clé privée au format PEM (PKCS#8) avec des permissions restrictives (0600)
	keyOut, err := os.OpenFile(KeyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	fmt.Printf("Successfully generated certificates:\n  Cert: %s\n  Key:  %s\n", CertPath, KeyPath)
	return nil
}
