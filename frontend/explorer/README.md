# Explorer

A block explorer for a Rivine-based chain

## Run it yourself

### Prerequisites
* Caddyserver
* A Rivine-based daemon


Make sure you have a Rivine-based daemon running with the explorer module enabled:
`<rivine-based daemon cli name> -M cgte`

Now start caddy from the `caddy` folder of this repository:
`caddy -conf Caddyfile.local`
and browse to http://localhost:2015
