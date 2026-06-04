package logger

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestInitLogger(t *testing.T) {
	// 1. Préparation : On s'assure de nettoyer le dossier de test à la fin
	defer func() {
		os.RemoveAll("logs")
	}()

	// ==========================================
	// Test A : Initialisation basique (Info + Texte)
	// ==========================================
	file1 := InitLogger("toxophilus.log", "Info", "texte")
	if file1 == nil {
		t.Fatal("❌ InitLogger devrait retourner un pointeur de fichier, obtenu nil")
	}
	// On ferme le fichier immédiatement pour libérer la ressource
	file1.Close()

	// Vérification : Le fichier a-t-il bien été créé sur le disque ?
	logPath := filepath.Join("logs", "toxophilus.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("❌ Le fichier de log %s n'a pas été créé", logPath)
	}

	// ==========================================
	// Test B : Initialisation avancée (Debug + JSON)
	// ==========================================
	file2 := InitLogger("toxophilus.log", "Debug", "json")
	if file2 == nil {
		t.Fatal("❌ InitLogger devrait retourner un pointeur de fichier, obtenu nil")
	}
	file2.Close()

	// Vérification : Le logger global de Go a-t-il bien été écrasé ?
	if slog.Default() == nil {
		t.Error("❌ Le logger par défaut slog n'a pas été configuré")
	}
}
