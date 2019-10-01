 # Integration with NBH backend systems

## App public key to customer

The Threefold app has a private/public key pair so a user can uniquely identify itself and sign requests to the public NBH API's.

At NBH side there is the notion of customers.

The big question is, **how do we link the two?**

The user can login into the onboarding UI using 3botlogin, at onboarding we can ask the customer if he already has a customerId from NBH. If so,  and approved,the user doesn't have to do KYC. 



## Customers vs addresses

At NBH side we talk about (KYC'ed) Customers. In the goldchain we have the notion of (authorized) addresses.

A User uses the Threefold app and identifies through the app. 
In the app, a goldchain wallet can be created and addresses can be created  from private keys. These addresses need to be authorized by NBH so they know which customer they belong to.

Open Questions:
- Where are the links between NBH customers, customer threefold app id's and authorized addresses stored?
   Options are
   - store it ourselves in bcdb
       - Is bcdb production ready?
   - store it in the NBH systems
       - How long will this take?
   
[A Jumpscale schema with the required data](customer_addresses.jsschema)

### Address Authorization
![Address authorization](Authorize_Addresses.svg)

If the KYC system updates the  BCDB with KYC valid date, the extra call to the KYC system can be dropped.

This is the logical flow. Technically it is more secure to put the process with the key for signing the authorization transaction on a seperate container that has no entrypoint from the outside.
### Remarks
- [Wallet remarks](wallet.md)
 ## Weight Account system

 The blockchain is master and there are reports available through an explorer ran on NBH systems to update the Weight account system balances.
 This way no api's have to be exposed.

 ![Blockchain to weight system](./WeightAccountUpdates.svg)

 Questions:
 - What needs to be in the reports? 
 - Does the weight account system hold the link between customers and it's addresses on the chain?
 - If not all transactions should be in the report, only a date parameter is sufficient to generate the report, else, it's a daterange.
 
## Gold acquisition

This is discussed in a [seperate topic](gold_acquisition.md).
