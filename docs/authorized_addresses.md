# Authorized addresses

For AML reasons, KYC needs to happen and only authorized users may own gold tokens.

From a chain perspective, this means that addresses need to be authorized before they can be used to own gold tokens.

## Implementation
The default [Rivine authorization extension](https://github.com/threefoldtech/rivine/tree/master/extensions/authcointx)  is used for this.

## Anonymization
It is recommended for authorized addresses to  dripple to the chain, if not, they are all linked together in the same transaction, coupling them to the same wallet.
