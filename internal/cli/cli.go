package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "didem",
		Short: "Distributed identity and email",
	}
)

func Execute() error {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "increase verbosity")
	rootCmd.PersistentFlags().String("addr", "127.0.0.1:8080", "address of daemon")
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		return errors.Wrap(err, "binding pflag")
	}

	if err := viper.BindPFlag("daemon_addr", rootCmd.PersistentFlags().Lookup("addr")); err != nil {
		return errors.Wrap(err, "binding pflag")
	}

	regCommands()

	return rootCmd.Execute()
}
