# Manifeste d'application SlashNode

> Statut : **spécification** (le moteur de registre/orchestration arrive dans une
> itération ultérieure). Ce document fige le format pour ne pas perdre les
> décisions déjà prises — notamment le bloc `inputs` (saisies utilisateur →
> variables d'environnement) et le mécanisme `exports`/`wiring`.

Une app du store, c'est un dossier :

```
bitcoind/
  slashnode-app.json     # manifeste (ce document)
  docker-compose.yml     # services Docker
  templates/*.tmpl       # templating de config (valeurs JSON → bitcoin.conf…)
  icon.svg
```

## Structure du manifeste

```jsonc
{
  "manifestVersion": 1,
  "id": "bitcoind",
  "name": "Bitcoin Core",
  "version": "27.0",
  "category": "bitcoin",
  "dependencies": [],

  // Saisies demandées à l'utilisateur au moment de l'installation.
  // Le formulaire d'install est généré à partir de cette liste, et chaque
  // valeur est injectée comme VARIABLE D'ENVIRONNEMENT dans le(s) conteneur(s).
  "inputs": [ /* voir ci-dessous */ ],

  "services": { /* … images, ports, volumes, healthchecks … */ },

  // Valeurs publiées dans le registre du démon, consommables par d'autres apps.
  "exports": { /* … */ }
}
```

## Le bloc `inputs` (saisies → variables d'env)

Chaque entrée décrit un champ de formulaire. La valeur saisie devient une
variable d'environnement (`key`) du conteneur.

```jsonc
"inputs": [
  {
    "key": "RPC_USER",            // nom EXACT de la variable d'env injectée
    "label": "Utilisateur RPC",
    "type": "text",               // type de champ (voir tableau)
    "required": true,
    "default": "satoshi",
    "placeholder": "ex. satoshi",
    "help": "Nom d'utilisateur pour l'API RPC.",
    "secret": false               // true => stocké chiffré, masqué dans l'UI
  },
  {
    "key": "ADMIN_EMAIL",
    "label": "Email administrateur",
    "type": "email",              // validation email côté formulaire
    "required": true,
    "help": "Pour les notifications du nœud."
  },
  {
    "key": "WALLET_PASSWORD",
    "label": "Mot de passe du wallet",
    "type": "password",           // champ masqué
    "required": true,
    "secret": true,               // n'est jamais réaffiché ni loggé
    "minLength": 12
  },
  {
    "key": "PRUNE_GB",
    "label": "Taille de prune (Go)",
    "type": "number",
    "default": 0,
    "min": 0
  },
  {
    "key": "NETWORK",
    "label": "Réseau",
    "type": "select",
    "options": ["mainnet", "testnet", "signet"],
    "default": "mainnet"
  },
  {
    "key": "ENABLE_TOR",
    "label": "Activer Tor",
    "type": "boolean",
    "default": true
  }
]
```

### Types de champ supportés

| `type`     | Rendu                      | Validation                     |
|------------|----------------------------|--------------------------------|
| `text`     | input texte                | `minLength`, `maxLength`, `pattern` |
| `email`    | input email                | format email                   |
| `password` | input masqué               | `minLength` ; impose `secret: true` |
| `number`   | input numérique            | `min`, `max`, `step`           |
| `textarea` | zone de texte multiligne   | `minLength`, `maxLength`        |
| `select`   | liste déroulante           | valeur ∈ `options`             |
| `boolean`  | interrupteur               | `true`/`false` → `"true"`/`"false"` |

### Règles d'injection

- La valeur est exposée au conteneur via `environment:` (la variable porte le
  nom `key`).
- `secret: true` (obligatoire pour `password`) → la valeur est stockée chiffrée
  dans `/var/lib/slashnode` et n'est jamais renvoyée à l'UI ni écrite dans les
  logs.
- Les valeurs non saisies mais avec `default` sont injectées telles quelles.
- `required: true` bloque l'installation tant que le champ est vide.

## `exports` / `wiring` (branchement automatique)

À l'installation, une app publie ses infos de connexion dans le registre du
démon (`exports`). Une app consommatrice les référence via `wiring` ; le démon
résout les références et les injecte au templating. Zéro config manuelle.

```jsonc
// bitcoind : publie
"exports": {
  "rpc.host": "bitcoind",
  "rpc.port": 8332,
  "rpc.user": "${input.RPC_USER}",
  "rpc.password": "${secret.RPC_PASSWORD}",
  "zmq.rawblock": "tcp://bitcoind:28332"
}
```

```jsonc
// lnd : consomme
"dependencies": ["bitcoind"],
"wiring": {
  "bitcoind.rpchost": "${bitcoind.exports.rpc.host}",
  "bitcoind.rpcuser": "${bitcoind.exports.rpc.user}",
  "bitcoind.rpcpass": "${bitcoind.exports.rpc.password}",
  "bitcoind.zmqpubrawblock": "${bitcoind.exports.zmq.rawblock}"
}
```

Références disponibles dans les valeurs : `${input.KEY}`, `${secret.KEY}`,
`${<app>.exports.<clé>}`.
