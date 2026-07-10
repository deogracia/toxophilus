package services

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"codeberg.org/go-pdf/fpdf"
	"github.com/deogracia/toxophilus/models"
	"github.com/spf13/viper"
)

var reBold = regexp.MustCompile(`\*\*(.*?)\*\*`)
var reItalic = regexp.MustCompile(`\*(.*?)\*`)

// markdownToHTML convertit les ** en <b> et les * en <i> pour le moteur HTML de GoFPDF
func markdownToHTML(text string) string {
	// On nettoie d'abord les retours à la ligne pour le moteur HTML de GoFPDF
	text = strings.ReplaceAll(text, "\r\n", "<br>")
	text = strings.ReplaceAll(text, "\n", "<br>")
	html := reBold.ReplaceAllString(text, "<b>$1</b>")
	html = reItalic.ReplaceAllString(html, "<i>$1</i>")
	return html
}

// GenerateContractPDF accepte en entrée le contrat ET les réglages dynamiques
// et génère le contrat au format PDF sur une seule page A4 à l'aide de GoFPDF.
func GenerateContractPDF(contract models.Contract, settings map[string]string) (string, error) {
	// Création du document PDF (Portrait, millimètres, format A4)
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.SetAutoPageBreak(true, 15)

	// Traducteur d'encodage pour afficher parfaitement les accents français et le symbole Euro (€)
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	// ==========================================
	// 1. CONFIGURATION DU PIED DE PAGE LÉGAL (Natif)
	// ==========================================
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 7)
		pdf.CellFormat(190, 3, tr(settings["pdf_footer_ligne1"]), "", 1, "C", false, 0, "")
		pdf.CellFormat(190, 3, tr(settings["pdf_footer_ligne2"]), "", 1, "C", false, 0, "")
		if settings["pdf_show_contact_footer"] == "true" {
			var contactParts []string
			if settings["club_address"] != "" {
				contactParts = append(contactParts, "Siège social : "+settings["club_address"])
			}
			if settings["club_website"] != "" {
				contactParts = append(contactParts, "Site Web : "+settings["club_website"])
			}
			if len(contactParts) > 0 {
				pdf.SetFont("Arial", "BI", 7)
				pdf.CellFormat(190, 3, tr(strings.Join(contactParts, " - ")), "", 1, "C", false, 0, "")
			}
		}
	})

	pdf.AddPage()

	// ==========================================
	// 2. EN-TÊTE DU CLUB (Image ou Texte)
	// ==========================================
	headerImagePath := settings["pdf_header_image"]

	if headerImagePath != "" {
		// Image d'en-tête (190 mm de large, hauteur calculée proportionnellement pour conserver le ratio d'aspect)
		info := pdf.RegisterImageOptions(headerImagePath, fpdf.ImageOptions{ImageType: "", ReadDpi: true})
		if info != nil && info.Width() > 0 {
			propHeight := (190.0 / info.Width()) * info.Height()
			pdf.ImageOptions(headerImagePath, 10, 10, 190, propHeight, false, fpdf.ImageOptions{ImageType: "", ReadDpi: true}, 0, "")
			pdf.SetY(10 + propHeight + 4)
		} else {
			pdf.ImageOptions(headerImagePath, 10, 10, 190, 22, false, fpdf.ImageOptions{ImageType: "", ReadDpi: true}, 0, "")
			pdf.SetY(36)
		}
	} else {
		pdf.SetY(10)
		pdf.SetFont("Arial", "B", 18)
		pdf.CellFormat(190, 9, tr(settings["pdf_club_name"]), "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "I", 11)
		pdf.CellFormat(190, 5, tr(settings["pdf_club_subtitle"]), "", 1, "C", false, 0, "")
		pdf.SetY(pdf.GetY() + 1)
	}
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.SetY(pdf.GetY() + 3)

	// ==========================================
	// 3. LOCATAIRE & CONTACTS
	// ==========================================
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(190, 5, tr("INFORMATIONS DU LOCATAIRE"), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(95, 4, tr(fmt.Sprintf("Nom / Prénom : %s %s", contract.Member.Nom, contract.Member.Prenom)), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, tr(fmt.Sprintf("N° Licence : %s", contract.Member.CodeAdherent)), "", 1, "L", false, 0, "")

	pdf.CellFormat(95, 4, tr(fmt.Sprintf("Adresse : %s, %s %s", contract.Member.StreetAddress, contract.Member.PostalCode, contract.Member.AddressLocality)), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, tr(fmt.Sprintf("Email : %s  -  Tél : %s", contract.Member.Email, contract.Member.Telephone)), "", 1, "L", false, 0, "")

	if contract.Member.IsMinor() {
		pdf.SetFont("Arial", "BI", 8)
		if contract.Member.NeedsParentalAuthorization() {
			pdf.CellFormat(190, 4, tr(fmt.Sprintf("Représentant légal (Mineur) : %s %s (%s) - Tél : %s - Email : %s",
				contract.Member.ParentPrenom, contract.Member.ParentNom, contract.Member.ParentRelation,
				contract.Member.ParentTelephone, contract.Member.ParentEmail)), "", 1, "L", false, 0, "")
		} else if contract.Member.EstEmancipe {
			pdf.CellFormat(190, 4, tr(fmt.Sprintf("Statut : Mineur émancipé (Décision : %s)", contract.Member.ReferenceEmancipation)), "", 1, "L", false, 0, "")
		}
	}
	pdf.SetY(pdf.GetY() + 1)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.SetY(pdf.GetY() + 2)

	// ==========================================
	// 4. MATÉRIEL MIS À DISPOSITION & PÉRIODE
	// ==========================================
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(190, 5, tr("MATÉRIEL MIS À DISPOSITION & PÉRIODE"), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)

	hasEquipment := false
	if contract.Riser.NumeroSerie != "" {
		hasEquipment = true
		pdf.CellFormat(190, 4, tr(fmt.Sprintf("Poignée : %s %s (N° %s)", contract.Riser.Marque, contract.Riser.Modele, contract.Riser.NumeroSerie)), "", 1, "L", false, 0, "")
	}
	if contract.Limb.NumeroSerie != "" {
		hasEquipment = true
		pdf.CellFormat(190, 4, tr(fmt.Sprintf("Branches : %s %s - %s / %s lbs (N° %s)", contract.Limb.Marque, contract.Limb.Modele, contract.Limb.Taille, contract.Limb.Puissance, contract.Limb.NumeroSerie)), "", 1, "L", false, 0, "")
	}
	if contract.Accessoires != "" {
		hasEquipment = true
		pdf.CellFormat(190, 4, tr(fmt.Sprintf("Accessoires : %s", contract.Accessoires)), "", 1, "L", false, 0, "")
	}
	if !hasEquipment {
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(190, 4, tr("Aucun arc spécifique. Accessoires uniquement."), "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 8)
	}

	pdf.SetFont("Arial", "BI", 8)
	pdf.CellFormat(190, 4, tr(fmt.Sprintf("Période de location : Du %s au %s", contract.DateDebut.Format("02/01/2006"), contract.DateFin.Format("02/01/2006"))), "", 1, "L", false, 0, "")
	if contract.Commentaire != "" {
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(190, 4, tr(fmt.Sprintf("Observations : %s", contract.Commentaire)), "", 1, "L", false, 0, "")
	}
	pdf.SetY(pdf.GetY() + 1)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.SetY(pdf.GetY() + 2)

	// ==========================================
	// 5. CONDITIONS FINANCIÈRES
	// ==========================================
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(190, 5, tr("CONDITIONS FINANCIÈRES"), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(95, 4, tr(fmt.Sprintf("Montant de la location : %.2f €", contract.MontantLocation)), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, tr(fmt.Sprintf("Montant de la caution : %.2f €", contract.MontantCaution)), "", 1, "L", false, 0, "")
	pdf.SetY(pdf.GetY() + 1)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.SetY(pdf.GetY() + 2)

	// ==========================================
	// 6. CONDITIONS DE LOCATION (Rendu HTML Natif)
	// ==========================================
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(190, 5, tr("CONDITIONS DE LOCATION"), "", 1, "L", false, 0, "")

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

	html := pdf.HTMLBasicNew()
	for _, c := range clauses {
		valeur := settings[c.Key]
		if valeur != "" {
			pdf.SetFont("Arial", "B", 8)
			pdf.CellFormat(190, 4, tr(c.Title), "", 1, "L", false, 0, "")
			pdf.SetFont("Arial", "", 8)

			// Conversion automatique et parfaite du Markdown vers le HTML simple de GoFPDF
			htmlText := tr(markdownToHTML(valeur))
			html.Write(3.5, htmlText)
			pdf.Ln(4.5) // Saut de ligne typographique aéré entre deux articles
		}
	}
	pdf.SetY(pdf.GetY() + 1)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.SetY(pdf.GetY() + 2)

	// ==========================================
	// 7. SIGNATURES
	// ==========================================
	dateCreation := contract.CreatedAt
	if dateCreation.IsZero() {
		dateCreation = contract.DateDebut
	}
	dateStr := dateCreation.Format("02/01/2006")

	clubCity := settings["club_address_locality"]
	if clubCity == "" {
		clubCity = "________________________"
	}

	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(95, 4, tr(fmt.Sprintf("Fait à : %s, le : %s", clubCity, dateStr)), "", 0, "L", false, 0, "")
	pdf.CellFormat(95, 4, tr(fmt.Sprintf("Fait à : %s, le : %s", clubCity, dateStr)), "", 1, "L", false, 0, "")
	pdf.Ln(1)

	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(95, 4, tr("Signature du Club :"), "", 0, "L", false, 0, "")
	if contract.Member.NeedsParentalAuthorization() {
		pdf.CellFormat(95, 4, tr("Signature du Représentant Légal :"), "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 7)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(95, 3, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(95, 3, tr(fmt.Sprintf("(Pour l'adhérent mineur %s %s)", contract.Member.Prenom, contract.Member.Nom)), "", 1, "L", false, 0, "")
		pdf.CellFormat(95, 3, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(95, 3, tr("(Précédée de la mention \"Lu et approuvé\")"), "", 1, "L", false, 0, "")
	} else if contract.Member.IsMinor() && contract.Member.EstEmancipe {
		pdf.CellFormat(95, 4, tr("Signature du Locataire (Mineur émancipé) :"), "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 7)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(95, 3, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(95, 3, tr("(Précédée de la mention \"Lu et approuvé\")"), "", 1, "L", false, 0, "")
	} else {
		pdf.CellFormat(95, 4, tr("Signature du Locataire :"), "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 7)
		pdf.SetTextColor(120, 120, 120)
		pdf.CellFormat(95, 3, "", "", 0, "L", false, 0, "")
		pdf.CellFormat(95, 3, tr("(Précédée de la mention \"Lu et approuvé\")"), "", 1, "L", false, 0, "")
	}
	pdf.SetTextColor(0, 0, 0)

	// Sauvegarde du document PDF
	pdfDir := viper.GetString("app.pdf_dir")
	if pdfDir == "" {
		pdfDir = "data/pdf"
	}
	_ = os.MkdirAll(pdfDir, 0750)

	filename := fmt.Sprintf("Contrat-N°%d-%s-%s-%s.pdf", contract.ID, contract.Member.Nom, contract.Member.Prenom, contract.DateDebut.Format("2006-01-02"))
	fullPath := filepath.Join(pdfDir, filename)
	err := pdf.OutputFileAndClose(fullPath)
	if err != nil {
		return "", err
	}

	return fullPath, nil
}
