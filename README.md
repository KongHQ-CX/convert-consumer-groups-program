# Convert Consumer Groups Program

Proprietary program to convert Consumer Groups (with Rate Limiting Policies)
from 2.8 format, to 3.4 "Scoped Plugin" format.

## Usage

1. Use [Kong deck](https://github.com/kong/deck) to dump out the gateway
configuration for some workspace:

```sh
$ deck gateway dump --workspace finance -o finance.yaml
```

2. Run this converter:

```sh
$ ./convert-macos-arm64 -s finance.yaml -o finance-converted.yaml
```

3. Use deck to sync it back:

```sh
$ deck gateway sync --workspace finance finance-converted.yaml
```

Done!
