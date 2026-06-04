# toxophilus
Application de gestion de la location de matériel de tir à l'arc

* utilise par défaut sqlite;
* supportera à terme postgresql et mariaDB/MySQL;
* devrait être "légère";

# Environement

nécessite 
    * [just](https://github.com/casey/just)
    * go >= 1.26.2 pour la compilation
    * [Inkscape](https://inkscape.org/)
    * [ImageMagick](https://imagemagick.org)
    * [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
    * [gosec](github.com/securego/gosec)

Lancer sans générer un executable

* sur win11/powershell
```powershell
just run-dev
```
