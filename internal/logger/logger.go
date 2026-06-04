package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

// InitLogger configure le logger global de l'application
func InitLogger(filePath string, logLevel string, logFormat string) *os.File {
	// 1. Création du dossier de logs s'il n'existe pas
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0750); err != nil {
		log.Printf("Erreur lors de la création du dossier de logs : %v", err)
	}

	// 2. Ouverture ou création du fichier de log
	safePath := filepath.Clean(filePath)
	logFile := filepath.Join(logDir, safePath)
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		panic("Impossible de créer le fichier de log: " + err.Error())
	}

	// 3. Définition du niveau de log
	logLevelConverti := slog.Level(0)
	err = logLevelConverti.UnmarshalText([]byte(logLevel))
	if err != nil {
		log.Panic("Impossible de convertir " + logLevel + "en slog.Level.\n Err: " + err.Error())
	}

	// 4. Double sortie : Console (os.Stdout) ET Fichier (file)
	multiWriter := io.MultiWriter(os.Stdout, file)

	// 5. Choix du format (Texte classique ou JSON structuré)
	var handler slog.Handler
	options := &slog.HandlerOptions{
		Level: logLevelConverti,
	}

	if logFormat == "json" {
		handler = slog.NewJSONHandler(multiWriter, options)
	} else {
		handler = slog.NewTextHandler(multiWriter, options)
	}

	// 6. Remplacement du logger par défaut de Go
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return file // On retourne le fichier pour pouvoir le fermer proprement à l'arrêt
}
