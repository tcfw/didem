package cli

import (
	"context"
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
	rootCmd = &cobra.Command{
		Use:  "didem",
		RunE: run,
	}
)

func Execute() error {
	rootCmd.Flags().BoolP("verbose", "v", false, "increase verbosity")
	viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))

	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
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

	api := api.NewAPI(node)
	defer api.Shutdown(ctx)

	go func() {
		if err := api.ListenAndServe(&net.TCPAddr{Port: 8080}); err != nil {
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
