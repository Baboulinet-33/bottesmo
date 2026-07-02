# Tests

## Tests unitaires (Go)

- `internal/game/game_test.go` : Tests de la logique de jeu (validation des statuts de lettres, gestion des parties, etc.).
- `internal/game/game_test.go` : Tests multi-joueur (création de room, gestion des joueurs, séquence de mots, classement).
- `internal/dictionary/dictionary_test.go` : Tests de chargement du dictionnaire, sélection quotidienne et aléatoire.
- `internal/handlers/game_test.go` : Tests du endpoint de status /api/status.

## Tests E2E (Playwright)

- `bottesmo.spec.js` : 13 tests couvrant :
  1. Saisie au clavier et soumission.
  2. Comportement de la touche Retour arrière.
  3. Pré-remplissage des lettres trouvées à la ligne suivante.
  4. Saisie via clavier virtuel (clic).
  5. Effacement via le bouton Suppr du clavier virtuel.
  6. Coloration des touches du clavier après soumission.
  7. Parcours complet d'une partie gagnante.
  8. Affichage responsive (viewport mobile 375×667).
  9. Mode Daily : mot déterministe avec première lettre verrouillée.
  10. Accueil : trois boutons de mode (Daily, Solo, Multijoueur).
  11. Connexion joueur tardif avec wordGames présent.
  12. Double-clic sur démarrer (éviter 400 sur UI, bouton désactivé).
  13. Connexion SSE - content-type text/event-stream.
  14-26. Tests API multi-joueur.
  27. Tests GET /api/status (3 tests).
  28-33. Tests de toggle de thème.

- `multiplayer_restart.spec.js` : Tests multi-joueur couvrant :
  1. Création d'une room et vérification du code.
  2. Connexion SSE et réception d'événements.
  3. Parcours complet d'une partie multi-joueur (2 joueurs).
  4. Redémarrage de partie par le créateur.
