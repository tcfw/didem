package cli

func regCommands() {
	//Em
	emCmd.AddCommand(em_sendCmd)

	//P2P
	p2pCmd.AddCommand(p2p_peersCmd)

	//Root
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(p2pCmd)
	rootCmd.AddCommand(emCmd)
}
