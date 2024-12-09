/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"muskex/cmds/job/ops"

	"github.com/spf13/cobra"
)

// TimerMinerOrderIncomeCmd represents the TimerMinerOrderIncome command
var TimerMinerOrderIncomeCmd = &cobra.Command{
	Use:   "TimerMinerOrderIncome",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TimerMinerOrderIncome called")
		ops.TimerMinerOrderIncome()
	},
}

func init() {
	rootCmd.AddCommand(TimerMinerOrderIncomeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// TimerMinerOrderIncomeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// TimerMinerOrderIncomeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
