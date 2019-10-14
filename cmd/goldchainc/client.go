package main

import (
	gccli "github.com/nbh-digital/goldchain/pkg/client"
	"github.com/threefoldtech/rivine/pkg/client"
)

// CommandLineClient extend for commands
type CommandLineClient struct {
	*client.CommandLineClient
}

// NewCommandLineClient creates a new goldchain commandline client
func NewCommandLineClient(address, name, userAgent string) (*CommandLineClient, error) {
	rivCli := new(client.CommandLineClient)
	client, err := client.NewCommandLineClient(address, name, userAgent, &client.OptionalCommandLineClientCommands{
		CommandLineClient: rivCli,
		WalletCmd:         gccli.CreateWalletCmd(rivCli),
		AtomicSwapCmd:     gccli.CreateAtomicSwapCmd(rivCli),
	})
	if err != nil {
		return nil, err
	}
	return &CommandLineClient{
		CommandLineClient: client,
	}, nil
}
