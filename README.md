# Grimhold

Jeu d'exploration de donjon en mode texte, qui se joue dans le terminal. Tu explores
des régions maudites, tu tombes sur des ennemis, tu les combats avec des dés et tu
gardes la liste de tes victoires. Les monstres viennent en direct de l'API
[Open5e](https://open5e.com) (règles D&D 5e, sans clé d'API) et la carte du monde est
dans le fichier `world.json`.

Go 1.22 ou plus, bibliothèque standard uniquement, connexion internet nécessaire.

## Build et run

Depuis la racine du projet (le jeu lit `world.json` dans le dossier courant) :

```bash
go build -o grimhold ./cmd/grimhold
./grimhold
```

Ou directement : `go run ./cmd/grimhold`. Avec `--continue`, la dernière sauvegarde
est chargée au démarrage.

## Commandes

| Commande | Rôle |
|---|---|
| `help` | affiche l'aide |
| `map` | liste les régions, leur danger et leurs sorties |
| `travel <region>` | voyage vers une région voisine (`travel ashen marsh` fonctionne aussi) |
| `look` | décrit la région actuelle |
| `explore` | explore la région, avec un risque de rencontre |
| `fight` | joue un tour de combat |
| `flee` | tente de fuir |
| `rest` | soigne complètement, une fois par visite d'une région |
| `character` | affiche la fiche du personnage |
| `slain` | liste les monstres vaincus et leur nombre |
| `bestiary`, `next`, `prev` | parcourt les monstres de l'API page par page |
| `inspect <monster>` | affiche la fiche d'un monstre |
| `save`, `load` | sauvegarde / charge la partie (`save.json`) |
| `quit` | quitte le jeu |

## Comment jouer

1. Tu commences au carrefour (`Crossroads`), une région sans danger. `map` montre les
   régions et leur niveau de danger, `look` décrit l'endroit où tu es.
2. `travel <region>` te déplace, mais seulement vers une sortie de la région actuelle.
3. `explore` fait apparaître un ennemi dans les régions dangereuses. Plus le danger est
   élevé, plus l'ennemi risque d'être fort.
4. Pendant un combat, `fight` joue **un tour** : tu attaques (`d20 + bonus` contre
   l'armure de l'ennemi, puis dés de dégâts), puis l'ennemi riposte. Répète `fight`
   jusqu'à la fin, ou tente `flee` pour t'échapper.
5. Une victoire donne de l'XP, parfois de l'or, et ajoute le monstre à ta liste `slain`.
   Avec assez d'XP, tu montes de niveau (plus de vie max et de bonus d'attaque).
6. Si tu perds, tu te réveilles au carrefour avec 1 PV. `rest` te soigne, une seule
   fois par visite d'une région.
7. `save` enregistre la partie (impossible pendant un combat), `load` la recharge.
