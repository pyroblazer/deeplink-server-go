package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"deeplink-server/handlers"
	"deeplink-server/internal"

	"github.com/spf13/cobra"
)

var (
	expire string
	ctx    = context.Background()
)

var rootCmd = &cobra.Command{
	Use:   "deeplink",
	Short: "Custom Deep Link Server",
	Run: func(cmd *cobra.Command, args []string) {
		internal.InitRedis()
		handlers.StartServer()
	},
}

var createCmd = &cobra.Command{
	Use:   "create <code> <url>",
	Short: "Create short link",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		internal.InitRedis()
		code := args[0]
		url := args[1]
		key := "dl:" + code
		var err error
		if expire != "" {
			dur, err := time.ParseDuration(expire)
			if err != nil {
				log.Fatal("Invalid duration format")
			}
			err = internal.Rdb.Set(ctx, key, url, dur).Err()
		} else {
			err = internal.Rdb.Set(ctx, key, url, 0).Err()
		}
		if err != nil {
			log.Fatalf("Error saving: %v", err)
		}
		fmt.Println("Saved link: http://localhost:8080/" + code)
	},
}

func Execute() {
	createCmd.Flags().StringVar(&expire, "expire", "", "Set expiry e.g. 10m")
	rootCmd.AddCommand(createCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
