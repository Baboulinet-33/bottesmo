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
│   ├── handlers/
│   │   └── game.go
│   └── version/
│       └── version.go
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

### `internal/version`

Defines the application's Semantic Version (`MAJOR.MINOR.PATCH`). The canonical version is stored in `internal/version/version.go` as the `Version` variable (initial value `0.1.0`). This is the single source of truth for the application version — the `package.json` version belongs to the Playwright test suite only.

### `AGENTS.md`

Convention file for agent-level instructions. Located at the repository root, it tells AI agents how to bump the version on every application code change. Agents should read this file before making code changes.

### `internal/dictionary`

Charge les mots depuis `words.txt` (ou une source configurable via `DICT_WORDS_SOURCE`), permet de les filtrer par longueur, de tirer un mot aléatoire ou déterministe. Le dictionnaire sert uniquement pour choisir le mot cible (modes daily et solo) et pour valider la correspondance des lettres — la validation orthographique est déléguée au client.

Les fichiers dictionnaire peuvent optionnellement commencer par une ligne `DICT_VERSION=<nombre>` pour indiquer la version. Cette ligne est ignorée lors du chargement des mots et affichée dans les logs.

### Validation orthographique

La validation orthographique des mots saisis par le joueur est effectuée côté client via Typo.js (dictionnaire Hunspell français). Le serveur n'effectue plus de validation dictionnaire.

## Spécifications

Le dossier [`specs/`](../specs/) est la source de vérité pour les spécifications fonctionnelles. Les specs sont organisées par domaine fonctionnel (règles du jeu, modes, API, modèle de données, dictionnaire, frontend, configuration serveur, tests, multi-joueur). L'ancien fichier `game.md` a été supprimé. Tout ajout de fonctionnalité doit mettre à jour ou créer le fichier de spec correspondant.

## Deployment

Le chart Helm se trouve dans `helm/bottesmo/`. L'image Docker est publiée sur Docker Hub : `bnoleau/bottesmo:latest`.

### Commande d'installation

```bash
helm install bottesmo ./helm/bottesmo \
  --namespace bottesmo --create-namespace \
  --set ingress.hosts[0].host=bottesmo.example.com
```

### Contrainte 1 réplica — IMPORTANT

L'état multijoueur (rooms, joueurs connectés, flux SSE) est stocké **en mémoire dans le pod**. Scaler à N > 1 réplicas sans backend partagé causerait des **rooms fantômes**. La stratégie de déploiement est `Recreate` pour éviter deux pods actifs simultanément.

**Ne pas activer l'autoscaling (HPA) sans avoir d'abord externalisé l'état multijoueur.**

Voir `helm/bottesmo/README.md` pour la documentation complète.
