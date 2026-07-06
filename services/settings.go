package services

import (
	"errors"
	"log"

	"github.com/deogracia/toxophilus/database"
	"github.com/deogracia/toxophilus/models"
	"gorm.io/gorm"
)

// GetSetting récupère la valeur du paramètre `cle`
// S'il n'est pas défini, il retourne `defaultVal`
func GetSetting(cle string, defaultVal string) string {
	var setting models.Setting
	if err := database.DB.Where("cle = ?", cle).First(&setting).Error; err != nil {
		return defaultVal
	}
	return setting.Valeur
}

// SetSetting sauve en BDD le couple `cle`/`valeur`
func SetSetting(cle string, valeur string) error {
	var setting models.Setting
	result := database.DB.Where("cle = ?", cle).First(&setting)

	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return database.DB.Create(&models.Setting{Cle: cle, Valeur: valeur}).Error
	}
	setting.Valeur = valeur
	return database.DB.Save(&setting).Error
}

// InitDefaultSettings charge en base quelques valeurs par défaut
func InitDefaultSettings() {
	defaults := map[string]string{
		"pdf_clause_mise_disposition":       "Le club, dont les données figurent en bas de ce contrat, met à disposition de l'adhérent le matériel d'archerie désigné ci-dessus, révisé, propre et en parfait état de fonctionnement pour la durée convenue. Le Locataire déclare avoir examiné le matériel lors de sa remise, l'avoir testé sur le pas de tir, et reconnaît qu'il lui est confié en parfait état de fonctionnement et d'entretien.",
		"pdf_clause_conditions_utilisation": "Le matériel est loué à titre strictement personnel et ne peut en aucun cas être prêté ou cédé à un tiers. Le matériel loué reste la propriété exclusive et insaisissable du Club. L'adhérent s'engage à utiliser le matériel exclusivement pour la pratique du tir à l'arc dans le respect des consignes de sécurité fédérales. Il est seul responsable de l'utilisation de l'arc et des dommages matériels ou corporels causés à lui-même ou à des tiers du fait de ce matériel. L'utilisation du matériel en dehors des pas de tir agréés ou des compétitions officielles se fait sous l'entière responsabilité du Locataire.",
		"pdf_clause_entretien_reparations":  "L'adhérent prendra le plus grand soin du matériel confié et assurera son entretien courant. Toute anomalie ou casse doit être signalée immédiatement au club. Les réparations liées à l'usure normale sont à la charge du club ; celles consécutives à une négligence ou mauvaise utilisation seront facturées à l'adhérent.",
		"pdf_clause_depot_garantie":         "Un chèque de dépôt de garantie (non encaissé) est remis lors de la signature du contrat. Ce chèque sera restitué à l'adhérent après restitution complète, vérification et nettoyage de l'ensemble du matériel loué. En cas de perte, de vol, ou de dommages excédant l'usure normale, le Club retiendra sur cette caution les sommes nécessaires à la remise en état ou au remplacement à neuf du matériel (sur présentation de facture ou de devis). Si le montant des dégâts est supérieur à la caution, le Locataire s'engage à régler la différence.",
		"pdf_clause_duree_restitution":      "La location est accordée pour la durée fermement fixée au contrat. À l'échéance, l'adhérent s'engage à restituer l'intégralité du matériel propre et au complet. Tout retard non justifié pourra entraîner l'encaissement de la caution ou le refus de locations futures.",
	}

	for cle, val := range defaults {
		var existing models.Setting
		if database.DB.Where("cle = ?", cle).First(&existing).Error != nil {
			if err := SetSetting(cle, val); err != nil {
				log.Fatalf("❌ Impossible de mettre la valeur par défaut %s: %s. \nErreur: %v", cle, val, err)
			}
		}
	}
}
