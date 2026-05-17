$srcDir =  ".\images\sources"
$dest = ".\static\favicon.ico"
$svgSrc = "$srcDir\Icone_inkscape.svg"
# Définition du tableau des tailles
$sizes = "16,32,48,128,256"

Write-Host "#########################################"
Write-Host "##                                     ##"
Write-Host "   Génération de $dest"
Write-Host "##                                     ##"
Write-Host "Source: $svgSrc"
Write-Host "Destination: $dest"
Write-Host "inkscape: $inkscapePath"
Write-Host "Tailles: $sizes"
Write-Host "##                                     ##"


# Assemblage des PNG en un fichier .ico
Write-Host "Génération du fichier $dest"
magick -background none "$svgSrc" -define icon:auto-resize="$sizes" "$dest"
Write-Host "##                                     ##"
Write-Host "#########################################"