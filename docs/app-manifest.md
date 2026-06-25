# SlashNode application manifest

A store app is a single JSON manifest. It carries a standard **docker-compose**
document in its `compose` field, so any docker-compose-only project is
compatible: paste its compose in, declare the user `inputs`, and you're done.

```
bitcoind/
  slashnode-app.json     # manifest (this document) — includes the compose doc
  icon.svg
```

## Manifest structure

```jsonc
{
  "manifestVersion": 1,
  "id": "bitcoind",
  "name": "Bitcoin Core",
  "version": "28.1",
  "category": "bitcoin",
  "dependencies": [],

  // Entries requested from the user at install time.
  // The install form is generated from this list, and each
  // value can be referenced as ${input.KEY} / ${secret.KEY} in `compose`.
  "inputs": [ /* see below */ ],

  // A standard docker-compose document (YAML). SlashNode templates the
  // ${input.X}/${secret.X}/${app.exports.key} references into it, then runs it
  // verbatim via `docker compose -f`. Any other ${VAR} is left untouched for
  // compose's own interpolation. Per-service image tags are selectable in the
  // UI (the version picker reads/overrides each service's `image:`), defaulting
  // to the latest stable release.
  "compose": "name: slashnode-bitcoind\nservices:\n  bitcoind:\n    image: \"bitcoin/bitcoin:28.1\"\n    …\n",

  // Optional config files, templated and written next to the compose file so it
  // can bind-mount them by relative path (e.g. ./config/lnd.conf:/path:ro).
  "configs": [ /* { "path": "/root/.lnd/lnd.conf", "content": "…" } */ ],

  // Values published in the daemon's registry, consumable by other apps.
  "exports": { /* … */ }
}
```

### Conventions

By convention each service joins the shared external network `slashnode` and
sets `container_name` equal to its service name, so apps reach each other by DNS
(e.g. an `exports` value of `"rpc.host": "bitcoind"`). Named volumes use stable
explicit names so data survives upgrades; a service can mount another app's
volume by referencing it as `external` (e.g. Electrs reading Bitcoin Core's
blocks).

## The `inputs` block (entries → env variables)

Each entry describes a form field. The entered value becomes an
environment variable (`key`) of the container.

```jsonc
"inputs": [
  {
    "key": "RPC_USER",            // EXACT name of the injected env variable
    "label": "RPC user",
    "type": "text",               // field type (see table)
    "required": true,
    "default": "satoshi",
    "placeholder": "e.g. satoshi",
    "help": "Username for the RPC API.",
    "secret": false               // true => stored encrypted, masked in the UI
  },
  {
    "key": "ADMIN_EMAIL",
    "label": "Administrator email",
    "type": "email",              // email validation on the form side
    "required": true,
    "help": "For node notifications."
  },
  {
    "key": "WALLET_PASSWORD",
    "label": "Wallet password",
    "type": "password",           // masked field
    "required": true,
    "secret": true,               // never displayed again nor logged
    "minLength": 12
  },
  {
    "key": "PRUNE_GB",
    "label": "Prune size (GB)",
    "type": "number",
    "default": 0,
    "min": 0
  },
  {
    "key": "NETWORK",
    "label": "Network",
    "type": "select",
    "options": ["mainnet", "testnet", "signet"],
    "default": "mainnet"
  },
  {
    "key": "ENABLE_TOR",
    "label": "Enable Tor",
    "type": "boolean",
    "default": true
  }
]
```

### Supported field types

| `type`     | Rendering                  | Validation                     |
|------------|----------------------------|--------------------------------|
| `text`     | text input                 | `minLength`, `maxLength`, `pattern` |
| `email`    | email input                | email format                   |
| `password` | masked input               | `minLength`; enforces `secret: true` |
| `number`   | numeric input              | `min`, `max`, `step`           |
| `textarea` | multiline text area        | `minLength`, `maxLength`        |
| `select`   | dropdown list              | value ∈ `options`             |
| `boolean`  | toggle                     | `true`/`false` → `"true"`/`"false"` |

### Injection rules

- The value is exposed to the container via `environment:` (the variable bears
  the name `key`).
- `secret: true` (mandatory for `password`) → the value is stored encrypted
  in `/var/lib/slashnode` and is never returned to the UI nor written to the
  logs.
- Values not entered but with a `default` are injected as-is.
- `required: true` blocks installation as long as the field is empty.

## `exports` + `dependencies` (automatic wiring)

On installation, an app publishes its connection info in the daemon's registry
(`exports`). Installing an app first auto-installs everything in its
`dependencies` (post-order), so a consuming app can reference a dependency's
exports directly anywhere in its `compose` (or `configs`) — no separate wiring
block. The daemon resolves the references at install time. Zero manual config.

```jsonc
// bitcoind: publishes
"exports": {
  "rpc.host": "bitcoind",
  "rpc.port": 8332,
  "rpc.user": "${input.RPC_USER}",
  "rpc.password": "${secret.RPC_PASSWORD}",
  "zmq.rawblock": "tcp://bitcoind:28332"
}
```

```jsonc
// lnd / mempool: consume directly inside their compose document
"dependencies": ["bitcoind"],
"compose": "…\n    environment:\n      CORE_RPC_HOST: \"${bitcoind.exports.rpc.host}\"\n      CORE_RPC_USERNAME: \"${bitcoind.exports.rpc.user}\"\n      CORE_RPC_PASSWORD: \"${bitcoind.exports.rpc.password}\"\n…"
```

References available in values: `${input.KEY}`, `${secret.KEY}`,
`${<app>.exports.<key>}`. Any other `${VAR}` in `compose` is left for docker
compose's own interpolation.
