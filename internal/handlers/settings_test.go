package handlers

import (
	"bytes"
	"errors"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/deogracia/toxophilus/models"
	"github.com/gin-gonic/gin"
)

// MockSettingRepository simule la base de données pour les réglages
type MockSettingRepository struct {
	settings map[string]string
}

func (m *MockSettingRepository) GetAll() ([]models.Setting, error) {
	var list []models.Setting
	for k, v := range m.settings {
		list = append(list, models.Setting{Cle: k, Valeur: v})
	}
	return list, nil
}

func (m *MockSettingRepository) GetByKey(key string) (*models.Setting, error) {
	val, ok := m.settings[key]
	if !ok {
		return nil, errors.New("record not found")
	}
	return &models.Setting{Cle: key, Valeur: val}, nil
}

func (m *MockSettingRepository) Save(setting *models.Setting) error {
	m.settings[setting.Cle] = setting.Valeur
	return nil
}

func (m *MockSettingRepository) SaveAll(settings map[string]string) error {
	for k, v := range settings {
		m.settings[k] = v
	}
	return nil
}

func TestProcessSettingsSave_ImageMimeValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Routeur et Handler de test
	mockRepo := &MockSettingRepository{settings: make(map[string]string)}
	handler := NewSettingHandler(mockRepo)

	r := gin.New()
	// On définit un template HTML minimal fictif pour éviter les paniques de Gin
	templ := template.Must(template.New("settings.html").Parse(`{{define "settings.html"}}mock{{end}}`))
	r.SetHTMLTemplate(templ)

	r.POST("/settings/save", handler.ProcessSettingsSave)

	t.Run("Rejet de fichier texte non autorisé", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// On simule l'envoi d'un fichier texte avec l'extension .png
		part, err := writer.CreateFormFile("header_image", "test_file.png")
		if err != nil {
			t.Fatalf("Impossible de créer le champ de fichier de formulaire : %v", err)
		}
		// Contenu binaire text/plain
		_, _ = part.Write([]byte("ceci est un simple fichier texte malveillant se faisant passer pour une image"))
		_ = writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/settings/save", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		r.ServeHTTP(w, req)

		// Le fichier texte doit être rejeté (HTTP 400 Bad Request)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Attendu 400 Bad Request, obtenu %d", w.Code)
		}
	})

	t.Run("Acceptation de format PNG valide", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("header_image", "test_image.png")
		if err != nil {
			t.Fatalf("Impossible de créer le champ de fichier de formulaire : %v", err)
		}
		// Magic bytes d'un fichier PNG réel complété à 512 octets
		pngHeader := make([]byte, 512)
		copy(pngHeader, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
		_, _ = part.Write(pngHeader)
		_ = writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/settings/save", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		r.ServeHTTP(w, req)

		// Un PNG valide doit rediriger l'utilisateur vers la page des paramètres (HTTP 303 See Other)
		if w.Code != http.StatusSeeOther {
			t.Errorf("Attendu 303 See Other pour un PNG valide, obtenu %d", w.Code)
		}
	})

	t.Run("Acceptation de format JPEG valide", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("header_image", "test_image.jpeg")
		if err != nil {
			t.Fatalf("Impossible de créer le champ de fichier de formulaire : %v", err)
		}
		// Magic bytes d'un fichier JPEG réel complété à 512 octets
		jpegHeader := make([]byte, 512)
		copy(jpegHeader, []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}) // Entête JFIF complète
		_, _ = part.Write(jpegHeader)
		_ = writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/settings/save", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		r.ServeHTTP(w, req)

		// Un JPEG valide doit rediriger l'utilisateur vers la page des paramètres (HTTP 303 See Other)
		if w.Code != http.StatusSeeOther {
			t.Errorf("Attendu 303 See Other pour un JPEG valide, obtenu %d", w.Code)
		}
	})
}
