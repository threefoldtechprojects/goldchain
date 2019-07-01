# goldchain

> :warning: **WARNING**: software is still in development. It is not yet meant for use in production.
> Standard network is disabled at the moment.

## devnet

You can run the chain easily on your computer to play with the software already.

First step is to launch a daemon from your console:

```
goldchaind --network devnet --no-bootstrap -Mgctwbe
```

the above launches a goldchain daemon on devnet, using no bootstrap
(meaning it doesn't try to connect to bootstrap nodes or wait for such nodes to know if you're sync or not),
enabling also the explorer module.

Once you have that you can recover the genesis wallet so you can start creating blocks and have money to spend:

```
goldchainc wallet recover --plain \
    --seed "carbon boss inject cover mountain fetch fiber fit tornado cloth wing dinosaur proof joy intact fabric thumb rebel borrow poet chair network expire else"
```

As this wallet is recovered as a plain wallet it does not have to be unlocked and is ready for use:

```
$ goldchainc wallet
Wallet status:
Encrypted, Unlocked
Confirmed Balance:   100006530 GFT
Locked Balance:      0 GFT
Unconfirmed Delta:   + 0 GFT
BlockStakes:         3000 BS
```

Please consult the `--help` menus of the `goldchainc` command and all its subcommands for more information on how to use the CLI.

### Using multiple wallets on the same machine

A single `goldchaind` daemon doesn't allow multiple wallets for the time being.
In order to have multiple wallets running on the same machine you therefore need
to run multiple `goldchaind` daemons, with each daemon:
  - using a unique persistent directory (either by starting each daemon from a different directory or
    by explicitly setting it using the `--persistent-dir` flag);
  - exposing itself using a unique port.
These different can manually be connected to one another using the `goldchainc gateway connect localhost:[port]` command.

### Authorized Address Management

Please consult the Rivine documentation about the Auth Coin Tx Extension for more information about this feature and its transactions:
<https://github.com/threefoldtech/rivine/blob/master/extensions/authcointx/README.md>

#### Authorization of an address using the CLI

To authorize and deauthorize addresses you can use the following command to create the transaction:

```
goldchainc wallet authcoin authaddresses \
    -e 0175e1a00548730d67ec1b46bc0fe469e7b9888cfab3c08548aaf900afaa52564520c537d665ca
    -e 01752fb52375a6b0521890673a9a901fce6c88e3e272613bf5eb0c467b064e773b6ce4c54a2931
    -d 017600553bc08234e99d0af79d775ad2da62ec7e0db51942dfa6861ea0b2bfd788a65c77e370d5
```

The above command authorizes the first 2 addresses using the shorthand alternative `-e` for the flag `-auth`,
while it deauthorizes the last address (`017600553bc08234e99d0af79d775ad2da62ec7e0db51942dfa6861ea0b2bfd788a65c77e370d5`)
using the shorthand alternative `-d` for the flag `-deauth`.

This command returns a freshly created transaction in JSON format which still has to be signed and submitted using
other wallet CLI commands. You can do it all in one step using shell piping:

```
goldchainc wallet send transaction "$(goldchainc wallet sign "$(goldchainc wallet authcoin authaddresses -e 0175e1a00548730d67ec1b46bc0fe469e7b9888cfab3c08548aaf900afaa52564520c537d665ca)")"
```

#### Transfer authorization powers

To transfer authorization power from the current condition to the new one, the following command can be executed:

```
goldchainc wallet send transaction "$(goldchainc wallet sign "$(goldchainc wallet authcoin updatecondition 0175e1a00548730d67ec1b46bc0fe469e7b9888cfab3c08548aaf900afaa52564520c537d665ca)")"
```

In order to transfer to a multi-signature multiple addresses can be defined (and optionally a sig count):

```
goldchainc wallet send transaction "$(goldchainc wallet sign "$(goldchainc wallet authcoin updatecondition 0175e1a00548730d67ec1b46bc0fe469e7b9888cfab3c08548aaf900afaa52564520c537d665ca 01752fb52375a6b0521890673a9a901fce6c88e3e272613bf5eb0c467b064e773b6ce4c54a2931 1)")"
```

### Minting

Please consult the Rivine documentation about the Minting Extension for more information about this feature and its transactions:
<https://github.com/threefoldtech/rivine/blob/master/extensions/minting/readme.md>

Coin minting works similarly to how to the authorized address management feature works.
Please consult the `--help` menu of the relevant commands using the following commands:

- Create a Minter Definition Transaction: `goldchainc wallet create minterdefinitiontransaction --help`
- Create a Coin Creation Transaction: `goldchainc wallet create coincreationtransaction --help`
