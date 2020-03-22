# vaultsync

Vault helper utility used to sync remote state for local development.

> NOTE: This should only be used for development purposes!

## Example

Below is a basic example of how to use `vaultsync`. Please note, it was written quickly and only for reference. Better documentation will be provided in upcoming changes.

### Config

The configuration file is the main source for `vaultsync` and is used to set the source and target Vault servers including their engines w/ secrets.

```
{
  "source_auth": {
    "address": "http://localhost:8300",
    "credentials": {
      "role_id": "<vault_role_id>,
      "secret_id": "<vault_secret_id>"
    },
    "method": "approle"
  },
  "source_policies_path": "./policies",
  "source_secrets": [
    {
      "engine": "kv",
      "mount": "example",
      "options": {
        "version": "1"
      },
      "paths": ["localhost/twitch_oauth_creds"]
    }
  ],
  "target_auth": {
    "address": "http://localhost:8200",
    "credentials": {
      "token": "<vault_token>"
    },
    "method": "token"
  },
  "target_auth_approles": [
    {
      "name": "api",
      "options": {
        "token_max_ttl": "30s",
        "token_policies": "api",
        "token_ttl": "30s"
      },
      "path": "approle"
    },
  ],
  "target_auth_methods": [
    {
      "options": {
        "type": "approle"
      },
      "output": "./approles",
      "path": "approle"
    }
  ],
}
```

### Run

To run the program, put a `config.json` in the same directory as the source code or use the CLI arg `--config=` to set your file at run-time.

```
vaultsync --config="./config.json"
```
