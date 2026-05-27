package services

import (
	"fmt"

	"github.com/deogracia/toxophilus/models"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

// GenerateContractPDF accepte le contrat ET les réglages dynamiques
func GenerateContractPDF(contract models.Contract, settings map[string]string) (string, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(15, 15, 15)

	// ==========================================
	// 1. PIED DE PAGE LÉGAL
	// ==========================================
	m.RegisterFooter(func() {
		m.Row(4, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_footer_ligne1"], props.Text{Size: 8, Style: consts.Italic, Align: consts.Center})
			})
		})
		m.Row(4, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_footer_ligne2"], props.Text{Size: 8, Style: consts.Italic, Align: consts.Center})
			})
		})
	})

	// ==========================================
	// 2. EN-TÊTE DU CLUB (Image ou Texte)
	// ==========================================
	headerImagePath := settings["pdf_header_image"]

	if headerImagePath != "" {
		// MODE IMAGE : Le club a uploadé son bandeau
		// On crée une ligne de 30 unités de haut (à ajuster selon les proportions de ton image)
		m.Row(30, func() {
			m.Col(12, func() {
				_ = m.FileImage(headerImagePath, props.Rect{
					Center:  true,
					Percent: 100, // L'image prendra toute la largeur disponible
				})
			})
		})
		m.Row(5, func() {}) // Petit espace sous l'image
	} else {
		// MODE TEXTE : Mode de secours si aucune image n'est configurée
		m.Row(15, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_club_name"], props.Text{Top: 3, Style: consts.Bold, Align: consts.Center, Size: 22})
			})
		})
		m.Row(8, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_club_subtitle"], props.Text{Style: consts.Italic, Align: consts.Center, Size: 14})
			})
		})
	}
	m.Line(10)

	// ==========================================
	// 3. LOCATAIRE & MATÉRIEL
	// ==========================================
	m.Row(8, func() {
		m.Col(12, func() { m.Text("INFORMATIONS DU LOCATAIRE", props.Text{Style: consts.Bold, Size: 11}) })
	})
	m.Row(6, func() {
		m.Col(6, func() { m.Text(fmt.Sprintf("Nom / Prénom : %s %s", contract.Member.Nom, contract.Member.Prenom)) })
		m.Col(6, func() { m.Text(fmt.Sprintf("N° Licence : %s", contract.Member.CodeAdherent)) })
	})
	m.Line(10)

	m.Row(8, func() {
		m.Col(12, func() { m.Text("MATÉRIEL MIS À DISPOSITION", props.Text{Style: consts.Bold, Size: 11}) })
	})
	if contract.Riser.NumeroSerie != "" {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Poignée : %s %s (N° %s)", contract.Riser.Marque, contract.Riser.Modele, contract.Riser.NumeroSerie))
			})
		})
	}
	if contract.Limb.NumeroSerie != "" {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Branches : %s %s - %s / %s (N° %s)", contract.Limb.Marque, contract.Limb.Modele, contract.Limb.Taille, contract.Limb.Puissance, contract.Limb.NumeroSerie))
			})
		})
	}
	m.Line(10)

	// ==========================================
	// 4. LES CLAUSES DYNAMIQUES
	// ==========================================
	if clauses := settings["pdf_clauses_location"]; clauses != "" {
		m.Row(40, func() {
			m.Col(12, func() {
				m.Text(clauses, props.Text{Size: 9})
			})
		})
	}

	// ==========================================
	// 5. SIGNATURES
	// ==========================================
	m.Row(30, func() {
		m.Col(6, func() { m.Text("Signature du Club :", props.Text{Style: consts.Bold}) })
		m.Col(6, func() {
			m.Text("Signature du Locataire :", props.Text{Style: consts.Bold})
			m.Text("(Précédée de la mention \"Lu et approuvé\")", props.Text{Top: 5, Size: 8, Color: color.NewWhite()})
		})
	})

	filename := fmt.Sprintf("contrat_%d.pdf", contract.ID)
	err := m.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
