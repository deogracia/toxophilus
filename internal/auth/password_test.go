package auth

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	password := "MonSuperMotDePasse123!"

	// 1. Test de la génération du hash
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Erreur inattendue lors du hachage : %v", err)
	}

	if hash == "" || hash == password {
		t.Error("Le hash est vide ou n'a pas été chiffré")
	}

	// 2. Test de la vérification avec le BON mot de passe
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash a retourné false pour un mot de passe correct")
	}

	// 3. Test de la vérification avec un MAUVAIS mot de passe
	mauvaisPassword := "MauvaisMotDePasse!"
	if CheckPasswordHash(mauvaisPassword, hash) {
		t.Error("CheckPasswordHash a retourné true pour un mot de passe incorrect")
	}
}
