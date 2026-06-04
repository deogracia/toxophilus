Get-ChildItem -File -Recurse | Where-Object {
    $_.Extension -notin @('.exe','.db','.jpg','.jpeg','.png','.ico','.svg','.pdf') -and
    $_.FullName -notmatch '\\\.git\\' -and
    $_.FullName -notmatch '\\dist\\' -and
    $_.Name -ne 'tout_mon_code.txt' -and
    $_.Name -ne 'arborescence.txt'
} | ForEach-Object {
    "`n=========================================`n--- FILE: $($_.FullName.Replace($PWD.Path + '\', '').Replace('\', '/')) ---`n=========================================`n" + (Get-Content $_.FullName -Raw)
} | Set-Content tout_mon_code.txt -Encoding UTF8
