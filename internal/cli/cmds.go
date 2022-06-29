package cli

func regCommands() {
	//P2P
	p2pCmd.AddCommand(p2p_peersCmd)

	//Root
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(p2pCmd)

}
