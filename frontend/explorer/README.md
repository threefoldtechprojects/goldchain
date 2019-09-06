# Explorer

A block explorer for GoldChain

## Run it yourself

### Prerequisites
* Caddyserver
* GoldChain daemon (`goldchaind`)


Make sure you have `goldchaind` (the GoldChain daemon) running with the explorer module enabled:
`goldchaind -M cgte`

Now start caddy from the `caddy` folder of this repository:
`caddy -conf Caddyfile.local`
and browse to http://localhost:2015
