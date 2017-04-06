package cmd

import (
	// pewpew "github.com/bengadbois/pewpew/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var benchCmd = &cobra.Command{
	Use:   "bench URL...",
	Short: "Run benchmark tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	RootCmd.AddCommand(benchCmd)
	benchCmd.Flags().IntP("begin", "b", 5, "Beginning requests per second.")
	viper.BindPFlag("begin", benchCmd.Flags().Lookup("begin"))

	benchCmd.Flags().IntP("end", "e", 15, "Ending requests per second.")
	viper.BindPFlag("end", benchCmd.Flags().Lookup("end"))

	benchCmd.Flags().IntP("interval", "i", 5, "How many requests to add per round until the end.")
	viper.BindPFlag("interval", benchCmd.Flags().Lookup("interval"))

	benchCmd.Flags().StringP("duration", "d", "10s", "How long each round lasts.")
	viper.BindPFlag("duration", benchCmd.Flags().Lookup("duration"))

	benchCmd.Flags().String("cooldown", "1s", "How long to pause between rounds.")
	viper.BindPFlag("cooldown", benchCmd.Flags().Lookup("cooldown"))
}
