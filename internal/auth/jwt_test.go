package auth

import (
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	// Préparation : on force une clé secrète dans Viper pour le test
	viper.Set("app.secret_key", "cle_de_test_tres_secrete")

	// On nettoie Viper à la fin du test pour ne pas polluer les autres tests
	defer viper.Reset()

	userID := uint(42)

	// 1. Génération du token
	tokenString, err := GenerateToken(userID)
	if err != nil {
		t.Fatalf("Erreur lors de la génération du JWT : %v", err)
	}
	if tokenString == "" {
		t.Error("Le token généré est vide")
	}

	// 2. Validation d'un token correct
	claims, err := ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("Erreur inattendue lors de la validation : %v", err)
	}

	// 3. Vérification du contenu du token (les claims)
	if claims.UserID != userID {
		t.Errorf("Attendu userID %d, obtenu %d", userID, claims.UserID)
	}

	// Vérification de l'expiration (doit être dans le futur)
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Le token généré est déjà expiré")
	}

	// 4. Test avec un token totalement bidon
	_, err = ValidateToken("ceci.n_est.pas_un_token_valide")
	if err == nil {
		t.Error("La validation aurait dû échouer avec un token invalide")
	}
}

func TestJWTWithoutSecretKey(t *testing.T) {
	// On s'assure que la clé est vide
	viper.Reset()

	_, err := GenerateToken(1)
	if err == nil {
		t.Error("Générer un token sans clé secrète configurée devrait renvoyer une erreur")
	}
}
