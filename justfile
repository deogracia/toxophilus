# --- ARCHERIE CLUB JUSTFILE ---

# On définit powershell spécifiquement pour windows.
set windows-shell := ["powershell.exe","-NoProfile","-command"]

# Cible par défaut si on tape juste `just`
default:
	@just --list

# Lance tous les tests unitaires du projet
test:
	go test ./...

# Lance les tests avec plus de détails (verbose)
test-v:
	go test -v ./...

# Génère et affiche la couverture de code dans le navigateur
coverage:
	go test -coverprofile coverage.out ./...
	go tool cover -html coverage.out
	@rm coverage.out # Nettoie le fichier temporaire

# --- BONUS POUR PLUS TARD ---

# Compile l'application
build:
	go build -o archerie-app main.go

# Lance l'application en local
run:
	go run main.go