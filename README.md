# Bottesmo

Jeu de mots inspiré de Wordle, en français.

Deux modes de jeu :
- **Mot du jour** : un mot identique pour tous les joueurs, changé chaque jour.
- **Solo** : un mot aléatoire à chaque partie.

## Lancer le serveur

```bash
go run ./cmd/server
```

Le serveur écoute sur le port `3102` par défaut. Pour utiliser un autre port :

```bash
PORT=3103 go run ./cmd/server
```

Ouvrir http://localhost:3102 dans un navigateur.

## Lancer les tests

```bash
go test ./... -v
```

## Modifier le dictionnaire

Le dictionnaire est contenu fichier `internal/dictionary/words.txt`.

Le dictionnaire est un fichier texte avec un mot par ligne, en majuscules :

```
ABRITE
ACCORD
ACTION
```

Deux fichiers dictionnaire sont utilisés :
- **`words.txt`** : les mots à deviner (tirés aléatoirement ou quotidiennement).
- **`words_full.txt`** : tous les mots français valides (utilisé pour valider les propositions des joueurs). Pour l'instant, c'est une copie de `words.txt`.

Vous pouvez remplacer `words_full.txt` par un dictionnaire français complet (ex. liste de mots du Scrabble, Hunspell, etc.) sans rien changer au code.

Après modification, redémarrer le serveur pour prendre en compte les changements.

Les mots sont automatiquement convertis en majuscules au chargement. Les lignes vides sont ignorées.

## Spécifications

Les spécifications fonctionnelles sont organisées par domaine dans le dossier [`specs/`](specs/) :

- [Règles du jeu](specs/game-rules.md)
- [Modes de jeu](specs/game-modes.md)
- [API](specs/api.md)
- [Modèle de données](specs/data-model.md)
- [Dictionnaire](specs/dictionary.md)
- [Variables d'environnement](specs/env.md)
- [Frontend](specs/frontend.md)
- [Configuration serveur](specs/server-config.md)
- [Tests](specs/testing.md)
- [Multi-joueur](specs/multiplayer.md)
