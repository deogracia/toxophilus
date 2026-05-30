# --- toxophilus JUSTFILE ---

# On définit powershell spécifiquement pour windows.
set windows-shell := ["powershell.exe","-NoProfile","-command"]
nom_application := "toxophilus"

# Cible par défaut si on tape juste `just`
default:
	@just --list

# Lance tous les tests unitaires du projet
[group('On my workstation')]
[group('tests & coverage')]
test:
	go test ./...

# Lance les tests avec plus de détails (verbose)
[group('On my workstation')]
[group('tests & coverage')]
test-v:
	go test -v ./...

# Génère et affiche la couverture de code dans le navigateur
[group('On my workstation')]
[group('tests & coverage')]
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

[group('On my workstation')]
[group('Building')]
build: (_ensure-dir "dist") generate-favicon
	@echo "🔨 Compilation des exécutables..."
	go build -o ./dist/{{nom_application}}-create-user{{extension}} ./cmd/create_user/main.go
	go build -o ./dist/{{exe_name}} ./cmd/server/main.go
	@cp config-example.toml ./dist/config-example.toml

# Lance l'application en local
[group('On my workstation')]
run:
	go run ./cmd/server/main.go

# Crée un administrateur (usage: just create-admin email mdp)
[group('On my workstation')]
create-admin email password:
	go run ./cmd/create_user/main.go {{email}} {{password}}

# Lance les tests puis l'application sans la compiler
[group('On my workstation')]
run-dev: test
	$env:TOXO_APP_SECRET_KEY="super, secretkey#"; $env:APP_ENV="development"; go run ./cmd/server/main.go

# Génère le fichier static/favicon.ico contenant les tailles 16,32,48,128 et 256
[group('On my workstation')]
generate-favicon:
	@powershell -NoProfile -ExecutionPolicy Bypass -File "./build-tools/generate-favicon.ps1"

[group('On my workstation')]
etat-projet:
	tree /F /A > arborescence.txt
	@powershell -NoProfile -ExecutionPolicy Bypass -File "./build-tools/all-files-in-one.ps1"
