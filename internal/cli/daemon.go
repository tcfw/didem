package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tcfw/didem/internal/api"
	"github.com/tcfw/didem/internal/node"
)

var (
	daemonCmd = &cobra.Command{
		Use:   "daemon",
		RunE:  runDaemon,
		Short: "run the daemon",
	}
)

func init() {
	daemonCmd.Flags().IntP("api-port", "p", 8080, "api port")
	viper.BindPFlag("api_port", daemonCmd.Flags().Lookup("api-port"))
}

func runDaemon(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := node.NewNode(ctx,
		node.WithDefaultOptions(ctx),
	)
	if err != nil {
		return errors.Wrap(err, "initing node")
	}

	errCh := make(chan error)

	go func() {
		if err := node.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	api, err := api.NewAPI(node)
	if err != nil {
		return err
	}
	defer api.Shutdown(ctx)

	go func() {
		fmt.Printf("Starting CLI API on %d", viper.GetInt("api_port"))
		if err := api.ListenAndServe(&net.TCPAddr{Port: viper.GetInt("api_port")}); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-waitExit(ctx):
		return node.Stop()
	}
}

func waitExit(ctx context.Context) <-chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	return sigs
}
