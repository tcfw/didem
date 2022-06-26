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
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	rootCmd.AddCommand(daemonCmd)

	return rootCmd.Execute()
}
