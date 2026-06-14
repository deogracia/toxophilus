package database

import (
	"log"

	"github.com/deogracia/toxophilus/models"

	"github.com/glebarez/sqlite"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect réalise la connection à la base de données en utilisant le bon driver
func Connect() {
	var err error
	var dialector gorm.Dialector

	driver := viper.GetString("database.driver")
	dsn := viper.GetString("database.dsn")

	switch driver {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	default:
		dialector = sqlite.Open(dsn)
	}

	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("❌ Impossible de se connecter à la BDD (%s): %v", driver, err)
	}

	log.Printf("✅ Connecté à la BDD (Moteur : %s)", driver)

	if err := DB.AutoMigrate(
		&models.User{}, &models.Member{}, &models.Limb{},
		&models.Riser{}, &models.Contract{}, &models.Setting{},
	); err != nil {
		log.Fatalf("🛑 Échec critique de la migration de la base de données : %v", err)
	}
}
