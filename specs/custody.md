# Custody fees

Storing and securing gold has a price, an adddress holding an amount of GFT pays the custody fee from the amount of GFT it holds. In practice this means that after a while the amount it can spend is less than the original amount, effectively having a spendable amount that degrades over time since the custody fee needs to be paid. When the address spends the GFT, only the spendable amount can be spent, the rest is either
- sent to an NBH digital account
- destroyed

NBH digital can at any point already claim the gold already reerved in custody fees. Given this, I'd suggest the custody fee GFT's are destroyed when an address uses it's spendable amount.
