# Règles du jeu

## Vue d'ensemble

Bottesmo est un jeu de devinettes de mots en français, inspiré de Wordle. Le joueur dispose de **6 tentatives** pour trouver un mot cible. Une lettre de départ est dévoilée pour amorcer la réflexion. Après chaque proposition, un code couleur indique la pertinence de chaque lettre.

## Mécanique de base

1. Le serveur choisit un **mot cible** (en français, en majuscules, de 6 à 10 lettres).
2. Le joueur voit uniquement la **première lettre** du mot cible (verrouillée, pré-remplie).
3. Le joueur a **6 tentatives** pour trouver le mot.
4. Chaque proposition doit :
   - Avoir la **même longueur** que le mot cible.
   - **Commencer par la première lettre** dévoilée.
   - Être un **mot français valide** (vérifié côté serveur).
5. Après chaque proposition, les lettres sont colorées selon leur statut.
6. Les lettres bien placées sont **verrouillées** et pré-remplies dans la ligne suivante.
7. La partie se termine quand le joueur trouve le mot ou épuise ses 6 tentatives.

## Système de feedback (code couleur)

| Couleur | Signification | Classe CSS | Contexte du jeu |
|---------|---------------|------------|-----------------|
| 🟩 Vert | Lettre **bonne** (bonne lettre, bonne position) | `.correct` | La lettre est correcte et bien placée. |
| 🟠 Ambre | Lettre **mal placée** (bonne lettre, mauvaise position) | `.present` | La lettre est dans le mot mais pas à cette position. |
| ⚫ Grise | Lettre **absente** (lettre pas dans le mot) | `.absent` | La lettre n'apparaît pas dans le mot cible. |

### Gestion des lettres doubles

L'algorithme suit la logique standard de Wordle pour les lettres dupliquées :
1. Les lettres **bien placées** (StatusCorrect) sont comptabilisées en premier.
2. Les occurrences restantes de chaque lettre sont ensuite attribuées aux positions **mal placées** (StatusPresent).
3. Les excédents sont marqués **absents**.

## Première lettre verrouillée

C'est la fonctionnalité signature qui distingue Bottesmo de Wordle standard :

1. La **première lettre** du mot cible est toujours pré-remplie et verrouillée dans **chaque ligne** de la grille.
2. Après chaque proposition, les lettres marquées `Correct` (bien placées) deviennent des lettres **"trouvées"**.
3. Ces lettres trouvées sont automatiquement pré-remplies et verrouillées dans la **ligne suivante**.
4. Les tuiles verrouillées ne peuvent pas être modifiées (la touche Retour arrière les ignore).

Ce mécanisme réduit progressivement l'espace de recherche et guide le joueur vers la solution.

## Conditions de victoire / défaite

| Issue | Condition | Message affiché |
|-------|-----------|-----------------|
| ✅ Victoire | La proposition correspond exactement au mot cible | "Bravo ! Vous avez trouvé le mot !" (vert) |
| ❌ Défaite | Après 6 propositions incorrectes | "Perdu ! Le mot était : {mot}" (rouge) |

Dans les deux cas, le mot cible est révélé et une option "Rejouer" est proposée.
