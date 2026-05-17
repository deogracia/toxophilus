$src =  ".\images\sources"
$dest = ".\static"
$svgSrc = "$src\Icone_inkscape.svg"
# Définition du tableau des tailles
$sizes = 16, 32, 48, 128, 256

# Chemin propre vers l'exécutable
$inkscapePath = Join-Path $env:ProgramFiles "inkscape\bin\inkscape.com"

Write-Host "#########################################"
Write-Host "##                                     ##"
Write-Host "   Génération de $dest\favicon.ico"
Write-Host "##                                     ##"
Write-Host "Source: $src"
Write-Host "Destination: $dest"
Write-Host "inkscape: $inkscapePath"
Write-Host "##                                     ##"

# Boucle pour générer chaque fichier PNG
foreach ($size in $sizes) {
    Write-Host "Génération du fichier $dest\$size.png"
    & $inkscapePath --export-type=png --export-filename="$dest\$size.png" -w $size -h $size "$svgSrc" *>$null
}

# Assemblage des PNG en un fichier .ico
Write-Host "Génération du fichier $dest\favicon.ico"
magick "$dest\16.png" "$dest\32.png" "$dest\48.png" "$dest\128.png" "$dest\256.png" -background none -colors 256 "$dest\favicon.ico"
Write-Host "##                                     ##"
Write-Host "#########################################"