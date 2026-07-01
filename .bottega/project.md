# Bottesmo

## Architecture

```
bottesmo/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── dictionary/
│   │   ├── dictionary.go
│   │   └── dictionary_test.go
│   ├── game/
│   │   ├── game.go
│   │   └── game_test.go
│   └── handlers/
│       └── game.go
├── web/
│   ├── templates/
│   │   ├── layout.html
│   │   └── game.html
│   └── static/
│       ├── lib/
│       │   └── typo/          # Typo.js + fr_FR Hunspell dictionary
│       ├── style.css
│       └── app.js
├── words.txt
└── go.mod
```

## Components

### `internal/dictionary`

Charge les mots depuis `words.txt`, permet de les filtrer par longueur, de tirer un mot aléatoire ou déterministe. Le dictionnaire sert uniquement pour choisir le mot cible (modes daily et solo) et pour valider la correspondance des lettres — la validation orthographique est déléguée au client.

### Validation orthographique

La validation orthographique des mots saisis par le joueur est effectuée côté client via Typo.js (dictionnaire Hunspell français). Le serveur n'effectue plus de validation dictionnaire.
