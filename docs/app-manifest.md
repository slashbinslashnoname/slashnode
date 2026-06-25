# SlashNode application manifest

> Status: **specification** (the registry/orchestration engine will arrive in a
> later iteration). This document freezes the format so we don't lose the
> decisions already made — in particular the `inputs` block (user entries →
> environment variables) and the `exports`/`wiring` mechanism.

A store app is a directory:

```
bitcoind/
  slashnode-app.json     # manifest (this document)
  docker-compose.yml     # Docker services
  templates/*.tmpl       # config templating (JSON values → bitcoin.conf…)
  icon.svg
```

## Manifest structure

```jsonc
{
  "manifestVersion": 1,
  "id": "bitcoind",
  "name": "Bitcoin Core",
  "version": "27.0",
  "category": "bitcoin",
  "dependencies": [],

  // Entries requested from the user at install time.
  // The install form is generated from this list, and each
  // value is injected as an ENVIRONMENT VARIABLE into the container(s).
  "inputs": [ /* see below */ ],

  "services": { /* … images, ports, volumes, healthchecks … */ },

  // Values published in the daemon's registry, consumable by other apps.
  "exports": { /* … */ }
}
```

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

## `exports` / `wiring` (automatic wiring)

On installation, an app publishes its connection info in the daemon's
registry (`exports`). A consuming app references them via `wiring`; the daemon
resolves the references and injects them into the templating. Zero manual config.

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
// lnd: consumes
"dependencies": ["bitcoind"],
"wiring": {
  "bitcoind.rpchost": "${bitcoind.exports.rpc.host}",
  "bitcoind.rpcuser": "${bitcoind.exports.rpc.user}",
  "bitcoind.rpcpass": "${bitcoind.exports.rpc.password}",
  "bitcoind.zmqpubrawblock": "${bitcoind.exports.zmq.rawblock}"
}
```

References available in values: `${input.KEY}`, `${secret.KEY}`,
`${<app>.exports.<key>}`.
