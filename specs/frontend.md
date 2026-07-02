# Frontend

## Stack

| Couche | Technologie |
|--------|-------------|
| Frontend | JavaScript vanilla, CSS vanilla |
| Templating | `html/template` (Go), pages rendues côté serveur |
| Stockage stats | `localStorage` |

## Thème et style

- **Double thème** : sombre (défaut) et clair, basculable via un bouton dans l'en-tête.
- **Bouton de thème** : intégré dans l'en-tête (flexbox `space-between`), stylisé sans fond ni bordure (`background: transparent; border: none`) avec un hover couleur accent (`var(--accent)`).
- **Persistance** : le choix du thème est sauvegardé dans `localStorage` (clé `bottesmo-theme`). À la première visite, le thème respecte la préférence système (`prefers-color-scheme`).
- **Mécanisme** : les couleurs sont définies via des **CSS custom properties** (`var(--…)`) dans `:root` (thème sombre) et `[data-theme="light"]` (surcharges claires).
- **Palette** : accent ambre `#de802b`, correct émeraude `#10b981`, présent ambre `#f59e0b`, absent gris `#6b7280`.
- **Police** : système (Segoe UI, Roboto, sans-serif).
- **Disposition** : centrée, largeur max 600px, flexbox.
- **Responsive** : point de rupture à 480px (taille réduite des tuiles et du clavier).

## Grille de jeu

- 6 rangées × N colonnes, espacement de 6px entre les tuiles.
- Tuiles de 52×52px (desktop) / 42×42px (mobile).
- Tuiles verrouillées : fond vert émeraude (`--correct`) avec bordure verte foncée.
- Curseur : bordure blanche clignotante (animation).
- Animation de soumission : `flip` (rotationX 90° puis retour) avec décalage de 100ms entre chaque tuile.

## Clavier virtuel

- Disposition **AZERTY** (française) :
  - Rangée 1 : `A Z E R T Y U I O P`
  - Rangée 2 : `Q S D F G H J K L M`
  - Rangée 3 : `[Entrée] W X C V B N [Suppr]`
- Touches de 32×48px (desktop) / 28×42px (mobile).
- Touches Entrée et Suppr plus larges.
- Après chaque proposition, les touches du clavier se colorent (priorité : Correct > Present > Absent).
- Interaction possible au clavier physique et au clic souris/tactile.

## Messages

- Erreur (mot invalide, problème réseau) : ambre (`var(--error)`).
- Victoire : vert émeraude (`var(--win)`).
- Défaite : ambre (`var(--accent)`).
