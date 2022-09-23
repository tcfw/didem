package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/user"
	"reflect"
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

		bootstrapperAddr = sync.NewTopic("bootstrap", reflect.TypeOf(""))
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
	viper.Set("p2p.bootstartPeers", []string{})

	usr, _ := user.Current()
	dir := usr.HomeDir
	dir = fmt.Sprintf("%s/.didem", dir)

	if err := os.MkdirAll(dir, 0644); err != nil {
		return err
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s/identities.yaml", dir), []byte{}, 0600); err != nil {
		return err
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s/didem.yaml", dir), []byte{}, 0600); err != nil {
		return err
	}

	var node *iNode.Node

	// if we're the first instance to signal, we'll become the LEADER.
	if seq == 1 {
		runenv.RecordMessage("i'm the leader.")
		numFollowers := runenv.TestInstanceCount - 1

		node, err = iNode.NewNode(ctx)
		if err != nil {
			return err
		}

		go node.ListenAndServe()
		defer node.Stop()

		h := node.P2P().Host()

		bootAddr := fmt.Sprintf("%s/p2p/%s", h.Addrs()[0].String(), h.ID().String())

		client.Publish(ctx, bootstrapperAddr, bootAddr)

		runenv.RecordMessage("bootstrap addr %s", bootAddr)

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
	} else {
		ch := make(chan string)
		client.MustSubscribe(ctx, bootstrapperAddr, ch)
		peer := <-ch

		runenv.RecordMessage("got bootstrap peer addr %s", peer)

		viper.Set("p2p.bootstartPeers", []string{peer})

		node, err = iNode.NewNode(ctx)
		if err != nil {
			return err
		}

		go node.ListenAndServe()
		defer node.Stop()
	}

	client.MustSignalEntry(ctx, readyState)

	// wait until the leader releases us.
	err = <-client.MustBarrier(ctx, releasedState, 1).C
	if err != nil {
		return err
	}

	return nil
}
