# Tests

## Tests unitaires (Go)

- `internal/game/game_test.go` : Tests de la logique de jeu (validation des statuts de lettres, gestion des parties, etc.).
- `internal/game/game_test.go` : Tests multi-joueur (création de room, gestion des joueurs, séquence de mots, classement).
- `internal/dictionary/dictionary_test.go` : Tests de chargement du dictionnaire, sélection quotidienne et aléatoire.

## Tests E2E (Playwright)

- `bottesmo.spec.js` : 9 tests couvrant :
  1. Saisie au clavier et soumission.
  2. Comportement de la touche Retour arrière.
  3. Pré-remplissage des lettres trouvées à la ligne suivante.
  4. Saisie via clavier virtuel (clic).
  5. Effacement via le bouton Suppr du clavier virtuel.
  6. Coloration des touches du clavier après soumission.
  7. Parcours complet d'une partie gagnante.
  8. Affichage responsive (viewport mobile 375×667).
  9. Mode Daily : mot déterministe avec première lettre verrouillée.

- `multiplayer_restart.spec.js` : Tests multi-joueur couvrant :
  1. Création d'une room et vérification du code.
  2. Connexion SSE et réception d'événements.
  3. Parcours complet d'une partie multi-joueur (2 joueurs).
  4. Redémarrage de partie par le créateur.
