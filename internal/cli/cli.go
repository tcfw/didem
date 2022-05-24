package cli

import (
	"context"
	"encoding/hex"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tcfw/didem/internal/node"
	"github.com/tcfw/didem/pkg/tx"
)

var (
	rootCmd = &cobra.Command{
		Use:  "didem",
		RunE: run,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := node.NewNode(
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

	id, err := node.Storage().PutTx(ctx, &tx.Tx{})
	if err != nil {
		return err
	}

	logrus.New().WithField("id", hex.EncodeToString(id)).Infof("stored TX")

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
