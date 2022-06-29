package cli

import (
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
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("daemon_addr", rootCmd.PersistentFlags().Lookup("addr"))

	regCommands()

	return rootCmd.Execute()
}
