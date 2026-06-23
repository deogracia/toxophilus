/* Scripts d'interaction dynamiques pour Toxophilus */

/**
 * switchTab gère la bascule d'onglets pour les composants de l'application
 * @param {string} tab - L'identifiant de l'onglet à activer ('risers' ou 'limbs')
 */
function switchTab(tab) {
    const risersContent = document.getElementById('content-risers');
    const limbsContent = document.getElementById('content-limbs');
    const risersTab = document.getElementById('tab-risers');
    const limbsTab = document.getElementById('tab-limbs');

    if (!risersContent || !limbsContent || !risersTab || !limbsTab) return;

    if (tab === 'risers') {
        risersContent.style.display = 'block';
        limbsContent.style.display = 'none';
        risersTab.className = ''; // Onglet actif (bouton plein de Pico)
        limbsTab.className = 'outline secondary'; // Onglet inactif
    } else {
        risersContent.style.display = 'none';
        limbsContent.style.display = 'block';
        risersTab.className = 'outline secondary';
        limbsTab.className = '';
    }
}
