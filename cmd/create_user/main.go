package main

import (
	"fmt"
	"log"
	"os"

	"github.com/deogracia/toxophilus/config"
	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/internal/auth"
	"github.com/deogracia/toxophilus/models"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: just create-admin <email> <password>")
		return
	}

	if err := config.LoadConfig(); err != nil {
		log.Fatalf("❌ Impossible de charger la configuration.\n Erreur: %v", err)
	}

	database.Connect()

	email := os.Args[1]
	password := os.Args[2]

	hashedPassword, _ := auth.HashPassword(password)
	user := models.User{Email: email, Password: hashedPassword}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Fatalf("Erreur : %v", err)
	}

	fmt.Printf("✅ Utilisateur %s créé avec succès !\n", email)
}
