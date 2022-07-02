package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	apipb "github.com/tcfw/didem/api"
	"github.com/tcfw/didem/internal/api"
	"github.com/tcfw/didem/internal/utils/logging"
)

var (
	p2pCmd = &cobra.Command{
		Use:   "p2p",
		Short: "P2P commands",
	}

	p2p_peersCmd = &cobra.Command{
		Use:   "peers",
		Short: "list peers",
		Run:   runP2PPeers,
	}
)

func runP2PPeers(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	api, err := api.NewClient()
	if err != nil {
		logging.WithError(err).Error("constructing client")
		return
	}

	res, err := api.P2P().Peers(ctx, &apipb.PeersRequest{})
	if err != nil {
		logging.WithError(err).Error("fetching peers")
		return
	}

	s, _ := json.Marshal(res.Peers)

	fmt.Printf("%s", s)

	return
}
