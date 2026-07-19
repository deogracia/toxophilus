# toxophilus
Application de gestion de la location de matériel de tir à l'arc

* utilise par défaut sqlite;
* supportera à terme postgresql et mariaDB/MySQL;
* devrait être "légère";

# Environnement

nécessite 
    * [just](https://github.com/casey/just)
    * go >= 1.26.2 pour la compilation
    * [Inkscape](https://inkscape.org/)
    * [ImageMagick](https://imagemagick.org)
    * [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
    * [gosec](github.com/securego/gosec)
    * [powerwhell](https://learn.microsoft.com/fr-fr/powershell/scripting/install/install-powershell-on-windows?view=powershell-7.6) > 7.6

## Lancer en développement (sans générer un exécutable)

* sur win11/powershell
```powershell
just run-dev
```

---

# Déploiement Continu (Alwaysdata)

L'application intègre des flux de livraison continue (CD) automatisés pour compiler et déployer vos binaires de production directement sur votre hébergement **Alwaysdata** de manière transparente, en gérant l'état du service.

## 🔑 Variables et Secrets à configurer

Pour que les pipelines de déploiement automatique de GitLab ou GitHub puissent se connecter et relancer votre service à distance, vous devez renseigner les variables d'environnement suivantes dans leurs interfaces respectives :

| Clé / Nom du Secret | Description | Exemple de Valeur |
| :--- | :--- | :--- |
| `SSH_PRIVATE_KEY` | Clé privée SSH autorisée sur votre compte Alwaysdata (à configurer dans l'onglet *Accès SSH* d'Alwaysdata). Elle permet le transfert sécurisé des fichiers via SCP. | `-----BEGIN OPENSSH PRIVATE KEY-----...` |
| `ALWAYSDATA_SSH_HOST` | L'hôte de connexion SSH de votre compte Alwaysdata. | `ssh-votrecompte.alwaysdata.net` |
| `ALWAYSDATA_ACCOUNT` | Le nom d'utilisateur (identifiant public) de votre compte Alwaysdata (il apparaît en haut à gauche de votre panel d'administration). | `votrecompte` |
| `ALWAYSDATA_SSH_PATH` | Le chemin du dossier distant où se trouve votre application sur le serveur SSH. | `/home/votrecompte/toxophilus` |
| `ALWAYSDATA_SERVICE_ID` | L'identifiant numérique de votre service configuré dans l'onglet *Services* d'Alwaysdata. | `25050` |
| `ALWAYSDATA_TOKEN` | Votre jeton d'API d'Alwaysdata (à générer dans *Compte > API*). | `2aa9d7da950447a6bd21ee2827978092` |

---

### 💻 1. Configuration sur GitHub (GitHub Actions)

Pour configurer les secrets sur votre dépôt public GitHub :
1. Sur GitHub, allez dans l'onglet **Settings** de votre dépôt.
2. Dans le menu latéral de gauche, cliquez sur **Secrets and variables > Actions**.
3. Cliquez sur le bouton vert **New repository secret**.
4. Ajoutez individuellement chacun des secrets répertoriés dans le tableau ci-dessus.

*Le déploiement se lancera manuellement à l'aide d'un simple clic depuis l'onglet **Actions > Deploy Release to Alwaysdata > Run workflow** en saisissant le tag que vous souhaitez déployer (ex: `v1.0.0`).*

---

### 🦊 2. Configuration sur GitLab (GitLab CI/CD)

Pour configurer les variables sur GitLab :
1. Sur GitLab, allez dans **Settings > CI/CD**.
2. Déroulez la section **Variables** et cliquez sur **Add variable**.
3. Renseignez la clé et la valeur pour chacune des variables répertoriées dans le tableau ci-dessus.
4. *Recommandation :* Cochez **Mask variable** (et **Protect variable** si nécessaire) pour sécuriser l'affichage de votre clé SSH privée et de votre token API.

*Le déploiement se déclenchera de manière autonome et sécurisée sous forme de job manuel (`when: manual`) dans l'onglet **CI/CD > Pipelines** lors de la création d'un tag de release (ex: `v1.0.0`).*

---

### 🛠️ 3. Déploiement en local (Justfile)

Vous pouvez également lancer un déploiement directement depuis votre poste à l'aide de la commande `just` :
```powershell
just deploy-tag <tag> <alwaysdata_token> <account_name> <service_id> <ssh_host> <ssh_path>
```
*(ex : `just deploy-tag v1.0.0 2aa9d7da95... votrecompte 25050 ssh-votrecompte.alwaysdata.net /home/votrecompte/toxophilus`)*
