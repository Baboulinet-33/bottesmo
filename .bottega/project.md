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

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server listening port | `3102` |
| `DICT_WORDS_SOURCE` | Source for target words dictionary (local path or HTTP(S) URL) | `internal/dictionary/words.txt` |
| `DICT_WORDS_FULL_SOURCE` | Source for full validation dictionary (local path or HTTP(S) URL) | `internal/dictionary/words_full.txt` |

Both `DICT_WORDS_SOURCE` and `DICT_WORDS_FULL_SOURCE` support auto-detection: if the value starts with `http://` or `https://` it is fetched via HTTP GET; otherwise it is treated as a local file path. At startup, the app logs each dictionary's version (if present), word count, and SHA256 hash.

## Components

### `internal/dictionary`

Charge les mots depuis `words.txt` (ou une source configurable via `DICT_WORDS_SOURCE`), permet de les filtrer par longueur, de tirer un mot aléatoire ou déterministe. Le dictionnaire sert uniquement pour choisir le mot cible (modes daily et solo) et pour valider la correspondance des lettres — la validation orthographique est déléguée au client.

Les fichiers dictionnaire peuvent optionnellement commencer par une ligne `DICT_VERSION=<nombre>` pour indiquer la version. Cette ligne est ignorée lors du chargement des mots et affichée dans les logs.

### Validation orthographique

La validation orthographique des mots saisis par le joueur est effectuée côté client via Typo.js (dictionnaire Hunspell français). Le serveur n'effectue plus de validation dictionnaire.
