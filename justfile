# --- toxophilus JUSTFILE ---

# On définit powershell spécifiquement pour windows.
set windows-shell := ["pwsh.exe","-NoProfile","-command"]
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
	go build -o ./dist/{{nom_application}}-cli{{extension}} ./cmd/toxophilus_cli/main.go
	go build -o ./dist/{{exe_name}} ./cmd/server/main.go
	@cp config-example.toml ./dist/config-example.toml

# Lance l'application en local
[group('On my workstation')]
run:
	go run ./cmd/server/main.go

# Crée un administrateur (usage: just create-admin email mdp)
[group('On my workstation')]
create-admin email password:
	go run ./cmd/toxophilus_cli/main.go create-admin {{email}} {{password}}

# Injecte des données factices de démonstration
[group('On my workstation')]
seed:
	go run ./cmd/toxophilus_cli/main.go seed

# Lance les tests puis l'application sans la compiler
[group('On my workstation')]
run-dev: generate-favicon test
	go run ./cmd/toxophilus_cli/main.go generate-certs
	$env:TOXO_APP_SECRET_KEY="dev_secret_key_must_be_at_least_32_chars_long_and_secure!"; $env:APP_ENV="development"; go run ./cmd/server/main.go

# Génère le fichier static/favicon.ico contenant les tailles 16,32,48,128 et 256
[group('On my workstation')]
generate-favicon:
	@powershell -NoProfile -ExecutionPolicy Bypass -File "./build-tools/generate-favicon.ps1"

[group('On my workstation')]
etat-projet:
	tree /F /A > arborescence.txt
	@powershell -NoProfile -ExecutionPolicy Bypass -File "./build-tools/all-files-in-one.ps1"

# Lance un audit complet de sécurité et de qualité
[group('Security & Quality')]
[group('On my workstation')]
audit:
	@echo "🔍 1. Vérification des dépendances avec govulncheck..."
	govulncheck ./...
	@echo ""
	@echo "🛡️ 2. Scan de sécurité du code avec gosec..."
	gosec -exclude-dir=dist ./...
	@echo ""
	@echo "🧹 3. Vérification des bonnes pratiques avec golangci-lint..."
	@echo " Todo..."
	# golangci-lint run

# Télécharge et déploie un tag de release spécifique (ex: just deploy-tag v1.0.0) sur Alwaysdata en gérant le service
[group('Deployment')]
[group('On my workstation')]
deploy-tag tag alwaysdata_token account_name service_id ssh_host ssh_path:
	@echo "📥 1. Téléchargement de l'archive du tag {{tag}} depuis GitHub..."
	@powershell -NoProfile -Command "New-Item -ItemType Directory -Force -Path ./dist | Out-Null"
	curl -s -L -o ./dist/toxophilus-linux-amd64.tar.gz "https://github.com/deogracia/toxophilus/releases/download/{{tag}}/toxophilus-linux-amd64.tar.gz"
	
	@echo "📦 2. Décompression locale du package..."
	tar -xzvf ./dist/toxophilus-linux-amd64.tar.gz -C ./dist/

	@echo "🛑 3. Mise en pause du service Alwaysdata..."
	curl -s -X PATCH "https://api.alwaysdata.com/v1/service/{{service_id}}/" \
		-u "{{alwaysdata_token}}:" \
		-H "Content-Type: application/json" \
		-d '{"is_disabled": true}'

	@echo "📤 4. Transfert des binaires extraits et de la configuration..."
	scp ./dist/toxophilus-linux-amd64 ./dist/toxophilus_cli-linux-amd64 ./dist/config-example.toml {{account_name}}@{{ssh_host}}:{{ssh_path}}/

	@echo "🚀 5. Relance du service Alwaysdata..."
	curl -s -X PATCH "https://api.alwaysdata.com/v1/service/{{service_id}}/" \
		-u "{{alwaysdata_token}}:" \
		-H "Content-Type: application/json" \
		-d '{"is_disabled": false}'
	@echo "✨ Déploiement du tag {{tag}} terminé avec succès !"
