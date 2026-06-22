package services

import (
	"fmt"
	"log"
	"strings"

	"github.com/deogracia/toxophilus/models"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

// renderRichText décode et affiche un texte avec une mise en forme par mot / ligne :
// - Si la ligne commence par - ou *, elle est affichée comme une puce d'indentation.
// - Les segments de texte entourés de ** sont affichés en Gras (ex: Le matériel est **personnel** ➡️ Le matériel est personnel).
func renderRichText(m pdf.Maroto, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			m.Row(2, func() {}) // Saut de ligne vide léger
			continue
		}

		isBullet := false
		// Extraction et détection des puces
		if strings.HasPrefix(trimmed, "- ") {
			trimmed = strings.TrimPrefix(trimmed, "- ")
			isBullet = true
		} else if strings.HasPrefix(trimmed, "* ") {
			trimmed = strings.TrimPrefix(trimmed, "* ")
			isBullet = true
		}

		// Calcul de la hauteur de ligne requise (limite de ~135 caractères par ligne à la taille 8)
		charCount := len(trimmed)
		if isBullet {
			charCount += 4
		}
		linesCount := charCount/135 + 1
		rowHeight := float64(linesCount) * 3.8
		if rowHeight < 4.5 {
			rowHeight = 4.5
		}

		m.Row(rowHeight, func() {
			m.Col(12, func() {
				currentLeft := 0.0
				if isBullet {
					m.Text("• ", props.Text{
						Size:  8,
						Style: consts.Bold,
						Color: color.Color{Red: 50, Green: 50, Blue: 50},
						Left:  1.0,
					})
					currentLeft = 4.0 // Décalage pour le texte de la puce
				}

				// Découpage par "**" pour alterner entre style Normal et style Gras
				parts := strings.Split(trimmed, "**")
				for idx, part := range parts {
					if part == "" {
						continue
					}

					style := consts.Normal
					textColor := color.Color{Red: 50, Green: 50, Blue: 50}
					if idx%2 == 1 {
						style = consts.Bold
						textColor = color.Color{Red: 10, Green: 10, Blue: 10}
					}

					m.Text(part, props.Text{
						Size:  8,
						Style: style,
						Color: textColor,
						Left:  currentLeft,
					})

					// Estimation de l'encombrement horizontal du segment imprimé :
					// Environ 1.35 unités de largeur par caractère à la taille de police 8.
					// Cela permet aux segments de s'écrire harmonieusement à la suite.
					currentLeft += float64(len(part)) * 1.32
				}
			})
		})
	}
}

// GenerateContractPDF accepte en entrée le contrat ET les réglages dynamiques
// et génère le contrat au format PDF sur une seule page A4.
func GenerateContractPDF(contract models.Contract, settings map[string]string) (string, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	// Marges compactées pour maximiser la hauteur verticale imprimable
	m.SetPageMargins(10, 10, 10)

	// ==========================================
	// 1. PIED DE PAGE LÉGAL (Compacté)
	// ==========================================
	m.RegisterFooter(func() {
		m.Row(3, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_footer_ligne1"], props.Text{Size: 7, Style: consts.Italic, Align: consts.Center})
			})
		})
		m.Row(3, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_footer_ligne2"], props.Text{Size: 7, Style: consts.Italic, Align: consts.Center})
			})
		})
		if settings["pdf_show_contact_footer"] == "true" {
			log.Println("pdf.go - affichage adresse physique et web demandée si renseignées")
			// Ligne d'adresse dédiée (si renseignée)
			if settings["club_address"] != "" {
				log.Println("pdf.go - affichage adresse physique demandée. adresse : " + settings["club_address"])
				m.Row(3, func() {
					m.Col(12, func() {
						m.Text("Siège social : "+settings["club_address"], props.Text{Size: 7, Style: consts.BoldItalic, Align: consts.Center})
					})
				})
			}
			// Ligne de site Web dédiée (si renseignée)
			if settings["club_website"] != "" {
				log.Println("pdf.go - affichage adresse web demandée. adresse : " + settings["club_website"])
				m.Row(3, func() {
					m.Col(12, func() {
						m.Text("Site Web : "+settings["club_website"], props.Text{Size: 7, Style: consts.BoldItalic, Align: consts.Center})
					})
				})
			}
		}
	})

	// ==========================================
	// 2. EN-TÊTE DU CLUB (Image ou Texte - Compacté)
	// ==========================================
	headerImagePath := settings["pdf_header_image"]

	if headerImagePath != "" {
		// Hauteur réduite à 22 pour économiser l'espace vertical
		m.Row(22, func() {
			m.Col(12, func() {
				_ = m.FileImage(headerImagePath, props.Rect{
					Center:  true,
					Percent: 100,
				})
			})
		})
		m.Row(3, func() {}) // Petit espace sous l'image
	} else {
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_club_name"], props.Text{Top: 2, Style: consts.Bold, Align: consts.Center, Size: 18})
			})
		})
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text(settings["pdf_club_subtitle"], props.Text{Style: consts.Italic, Align: consts.Center, Size: 11})
			})
		})
	}
	m.Line(3) // Ligne très fine pour préserver la hauteur

	// ==========================================
	// 3. LOCATAIRE, MATÉRIEL & PÉRIODE (Compacté)
	// ==========================================
	m.Row(6, func() {
		m.Col(12, func() { m.Text("INFORMATIONS DU LOCATAIRE", props.Text{Style: consts.Bold, Size: 9}) })
	})
	m.Row(5, func() {
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Nom / Prénom : %s %s", contract.Member.Nom, contract.Member.Prenom), props.Text{Size: 8})
		})
		m.Col(6, func() { m.Text(fmt.Sprintf("N° Licence : %s", contract.Member.CodeAdherent), props.Text{Size: 8}) })
	})
	m.Line(3)

	m.Row(6, func() {
		m.Col(12, func() { m.Text("MATÉRIEL MIS À DISPOSITION & PÉRIODE", props.Text{Style: consts.Bold, Size: 9}) })
	})

	hasEquipment := false
	if contract.Riser.NumeroSerie != "" {
		hasEquipment = true
		m.Row(5, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Poignée : %s %s (N° %s)", contract.Riser.Marque, contract.Riser.Modele, contract.Riser.NumeroSerie), props.Text{Size: 8})
			})
		})
	}
	if contract.Limb.NumeroSerie != "" {
		hasEquipment = true
		m.Row(5, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Branches : %s %s - %s / %s (N° %s)", contract.Limb.Marque, contract.Limb.Modele, contract.Limb.Taille, contract.Limb.Puissance, contract.Limb.NumeroSerie), props.Text{Size: 8})
			})
		})
	}
	if !hasEquipment {
		m.Row(5, func() {
			m.Col(12, func() {
				m.Text("Aucun arc spécifique. Accessoires uniquement.", props.Text{Size: 8, Style: consts.Italic})
			})
		})
	}

	// Affichage formel de la Période de Location
	m.Row(5, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Période de location : Du %s au %s", contract.DateDebut.Format("02/01/2006"), contract.DateFin.Format("02/01/2006")), props.Text{Style: consts.BoldItalic, Size: 8})
		})
	})
	m.Line(3)

	// ==========================================
	// 4. CONDITIONS FINANCIÈRES (Compactées)
	// ==========================================
	m.Row(6, func() {
		m.Col(12, func() { m.Text("CONDITIONS FINANCIÈRES", props.Text{Style: consts.Bold, Size: 9}) })
	})

	m.Row(5, func() {
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Montant de la location : %.2f €", contract.MontantLocation), props.Text{Size: 8})
		})
		m.Col(6, func() {
			m.Text(fmt.Sprintf("Montant de la caution : %.2f €", contract.MontantCaution), props.Text{Size: 8})
		})
	})

	m.Line(3)

	// ==========================================
	// 5. CONDITIONS DE LOCATION (Mise en forme riche)
	// ==========================================
	m.Row(6, func() {
		m.Col(12, func() { m.Text("CONDITIONS DE LOCATION", props.Text{Style: consts.Bold, Size: 9}) })
	})

	clauses := []struct {
		Title string
		Key   string
	}{
		{"1. Mise à disposition", "pdf_clause_mise_disposition"},
		{"2. Conditions d'utilisation", "pdf_clause_conditions_utilisation"},
		{"3. Entretien et réparations", "pdf_clause_entretien_reparations"},
		{"4. Dépôt de garantie (Caution)", "pdf_clause_depot_garantie"},
		{"5. Durée et restitution", "pdf_clause_duree_restitution"},
	}

	for _, c := range clauses {
		valeur := settings[c.Key]
		if valeur != "" {
			m.Row(4, func() {
				m.Col(12, func() {
					m.Text(c.Title, props.Text{Style: consts.Bold, Size: 8})
				})
			})
			renderRichText(m, valeur)
			m.Row(1, func() {}) // Micro-espace entre clauses
		}
	}
	m.Line(3)

	// ==========================================
	// 6. SIGNATURES (Compactées)
	// ==========================================
	// Hauteur ajustée à 16 unités pour éviter tout débordement sur la page 2
	m.Row(16, func() {
		m.Col(6, func() { m.Text("Signature du Club :", props.Text{Style: consts.Bold, Size: 8}) })
		m.Col(6, func() {
			m.Text("Signature du Locataire :", props.Text{Style: consts.Bold, Size: 8})
			m.Text("(Précédée de la mention \"Lu et approuvé\")", props.Text{Top: 4, Size: 7, Color: color.Color{Red: 120, Green: 120, Blue: 120}})
		})
	})

	filename := fmt.Sprintf("contrat_%d.pdf", contract.ID)
	err := m.OutputFileAndClose(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
}
