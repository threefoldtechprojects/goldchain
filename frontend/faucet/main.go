package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/nbh-digital/goldchain/pkg/config"
	"github.com/threefoldtech/rivine/extensions/authcointx"
	"github.com/threefoldtech/rivine/modules"
	"github.com/threefoldtech/rivine/pkg/daemon"
	"github.com/threefoldtech/rivine/types"

	gtypes "github.com/nbh-digital/goldchain/pkg/types"
)

type faucet struct {
	// cts is a cached version of daemon constants
	// caching here avoids requiring a call to the daemon even if it is local
	cts *modules.DaemonConstants
	// coinsToGive is the amount of coins given in a single transaction
	coinsToGive types.Currency
}

var (
	websitePort int
	httpClient  = &HTTPClient{
		RootURL:   "http://localhost:22110",
		Password:  "",
		UserAgent: daemon.RivineUserAgent,
	}
	coinsToGive uint64 = 300
)

func getDaemonConstants() (*modules.DaemonConstants, error) {
	var constants modules.DaemonConstants
	err := httpClient.GetAPI("/daemon/constants", &constants)
	if err != nil {
		return nil, err
	}
	return &constants, nil
}

func main() {
	log.Println("[INFO] Starting faucet")
	log.Println("[INFO] Loading daemon constants")
	cts, err := getDaemonConstants()
	if err != nil {
		panic(err)
	}

	f := faucet{
		cts:         cts,
		coinsToGive: config.GetTestnetGenesis().CurrencyUnits.OneCoin.Mul64(coinsToGive),
	}

	log.Println("[INFO] Faucet listening on port", websitePort)

	http.HandleFunc("/", f.requestFormHandler)
	http.HandleFunc("/request/tokens", f.requestTokensHandler)
	http.HandleFunc("/request/authorize", f.requestAuthorizationHandler)

	log.Println("[INFO] Faucet ready to serve")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", websitePort), nil))
}

func init() {
	flag.IntVar(&websitePort, "port", 2020, "local port to expose this web faucet on")
	flag.StringVar(&httpClient.Password, "daemon-password", httpClient.Password, "optional password, should the used daemon require it")
	flag.StringVar(&httpClient.RootURL, "daemon-address", httpClient.RootURL, "address of the daemon (with unlocked wallet) to talk to")
	flag.Uint64Var(&coinsToGive, "fund-amount", coinsToGive, "amount of coins to give per drip of the faucet")
	flag.Parse()

	// register tx versions for authentication
	_ = authcointx.NewPlugin(
		config.GetTestnetGenesisAuthCoinCondition(),
		gtypes.TransactionVersionAuthAddressUpdateTx,
		gtypes.TransactionVersionAuthConditionUpdateTx,
	)
}
