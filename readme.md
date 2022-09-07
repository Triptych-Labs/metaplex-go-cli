# Metaplex CLI

## Requirements
* `Go(lang)`

## Setup
```bash
$ git clone git@github.com:Triptych-Labs/metaplex-go-cli.git
$ cd metaplex-go-cli
$ go mod vendor
$ go run main.go <COMMAND>
```

## Operations

### `update_royalties_using_hashlist`
#### Usage
```bash
$ go run main.go update_royalties_using_hashlist <PATH_TO_KEYPAIR> <PATH_TO_HASHLIST> <ROYALTIES_AMOUNT>
```
#### Example
Set the royalties to `4%` of `./hashlist.json` mint addresses using keypair `./oracle.json`

```bash
$ go run main.go update_royalties_using_hashlist oracle.json hashlist.json 400
```
