package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/viper"
	"github.com/testground/sdk-go/network"
	"github.com/testground/sdk-go/run"
	"github.com/testground/sdk-go/runtime"
	"github.com/testground/sdk-go/sync"

	iNode "github.com/tcfw/didem/internal/node"
)

func main() {
	run.InvokeMap(map[string]interface{}{
		"tipset": runTipset,
	})
}

func runTipset(runenv *runtime.RunEnv) error {
	var (
		enrolledState = sync.State("enrolled")
		readyState    = sync.State("ready")
		releasedState = sync.State("released")
	)

	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	client := sync.MustBoundClient(ctx, runenv)
	defer client.Close()

	netclient := network.NewClient(client, runenv)
	runenv.RecordMessage("waiting for network initialization")

	// wait for the network to initialize; this should be pretty fast.
	netclient.MustWaitNetworkInitialized(ctx)
	runenv.RecordMessage("network initilization complete")

	seq := client.MustSignalEntry(ctx, enrolledState)

	runenv.RecordMessage("my sequence ID: %d", seq)

	ip, err := netclient.GetDataNetworkIP()
	if err != nil {
		return err
	}

	port := rand.Int()%1000 + 8712

	viper.Set("p2p.listeningAddrs", fmt.Sprintf("/ip4/%s/udp/%d/quic", ip.String(), port))

	node, err := iNode.NewNode(ctx)
	if err != nil {
		return err
	}

	go node.ListenAndServe()
	defer node.Stop()

	// if we're the first instance to signal, we'll become the LEADER.
	if seq == 1 {
		runenv.RecordMessage("i'm the leader.")
		numFollowers := runenv.TestInstanceCount - 1

		// let's wait for the followers to signal.
		runenv.RecordMessage("waiting for %d instances to become ready", numFollowers)
		err := <-client.MustBarrier(ctx, readyState, numFollowers).C
		if err != nil {
			return err
		}

		runenv.RecordMessage("the followers are all ready")
		runenv.RecordMessage("ready...")
		time.Sleep(1 * time.Second)
		runenv.RecordMessage("set...")
		time.Sleep(5 * time.Second)
		runenv.RecordMessage("go, release followers!")

		// signal on the 'released' state.
		client.MustSignalEntry(ctx, releasedState)
		return nil
	}

	client.MustSignalEntry(ctx, readyState)

	// wait until the leader releases us.
	err = <-client.MustBarrier(ctx, releasedState, 1).C
	if err != nil {
		return err
	}

	return nil
}
