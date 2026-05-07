# --- ARCHERIE CLUB JUSTFILE ---

# On définit powershell spécifiquement pour windows.
set windows-shell := ["powershell.exe","-NoProfile","-command"]
nom_application := "toxophilus"

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
# On fixe l'extension pour windows

extension := if os_family() == "windows" {".exe"} else {""}
exe_name := nom_application + extension

_ensure-dir target_dir:
	{{ if os() == "windows" { 
	"powershell -NoProfile -Command \"New-Item -ItemType Directory -Force -Path " + target_dir + " | Out-Null\"" 
	} else { 
	"mkdir -p " + target_dir 
	}}}

build: (_ensure-dir "dist")
	@echo "🔨 Compilation des exécutables..."
	go build -o ./dist/{{nom_application}}-create-user{{extension}} ./cmd/create_user/main.go
	go build -o ./dist/{{exe_name}} ./cmd/server/main.go
	@cp config-example.toml ./dist/config-example.toml

# Lance l'application en local
run:
	go run ./cmd/server/main.go

# Crée un administrateur (usage: just create-admin email mdp)
create-admin email password:
	go run ./cmd/create_user/main.go {{email}} {{password}}