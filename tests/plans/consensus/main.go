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

	"github.com/libp2p/go-libp2p-core/peer"
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
		signingKeys = runenv.StringArrayParam("signingKeys")
		genesis     = runenv.StringParam("genesis")

		enrolledState = sync.State("enrolled")
		readyState    = sync.State("ready")

		addrs = sync.NewTopic("addrs", reflect.TypeOf(""))
	)

	//check availablility of keys
	if len(signingKeys) != runenv.TestGroupInstanceCount {
		panic("not enough signing keys")
	}

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
	viper.Set("chain.genesis", genesis)
	viper.Set("chain.key", signingKeys[seq])

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

	node, err = iNode.NewNode(ctx)
	if err != nil {
		return err
	}

	go node.ListenAndServe()
	defer node.Stop()

	h := node.P2P().Host()

	bootAddr := fmt.Sprintf("%s/p2p/%s", h.Addrs()[0].String(), h.ID().String())

	client.Publish(ctx, addrs, bootAddr)

	runenv.RecordMessage("addr %d %s", seq, bootAddr)

	node, err = iNode.NewNode(ctx)
	if err != nil {
		return err
	}

	go node.ListenAndServe()
	defer node.Stop()

	ch := make(chan string)
	client.MustSubscribe(ctx, addrs, ch)

	for i := 0; i < runenv.TestInstanceCount-1; i++ {
		addr := <-ch

		addrInfo, err := peer.AddrInfoFromString(addr)
		if err != nil {
			runenv.RecordMessage("err parsing peer addr", err)
			continue
		}

		if err := node.P2P().Connect(ctx, *addrInfo); err != nil {
			runenv.RecordMessage("err connecting to peer %s: ", addrInfo.String(), err)
			continue
		}
	}

	client.MustSignalAndWait(ctx, readyState, runenv.TestGroupInstanceCount)

	return nil
}
