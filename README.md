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

Le fichier `words.txt` à la racine du projet est un lien symbolique vers `internal/dictionary/words.txt`. Vous pouvez modifier l'un ou l'autre.

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
