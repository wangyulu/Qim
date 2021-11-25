package mock

import (
	"context"
	"fmt"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

type StartOptions struct {
	protocol string
	addr     string
}

func NewClientCmd(ctx context.Context) *cobra.Command {
	opts := StartOptions{}

	cmd := &cobra.Command{
		Use:   "mock_cli",
		Short: "mock client ws",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := &ClientDemo{}

			if opts.protocol == "ws" && !strings.HasPrefix(opts.addr, "ws:") {
				opts.addr = fmt.Sprintf("ws://%s", opts.addr)
			}

			cli.Start(ksuid.New().String(), opts.protocol, opts.addr)

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol tcp or ws")
	cmd.PersistentFlags().StringVarP(&opts.addr, "addr", "a", ":9001", "server address")

	return cmd
}

func NewServerCmd(ctx context.Context) *cobra.Command {
	opts := StartOptions{}

	cmd := &cobra.Command{
		Use:   "mock_serv",
		Short: "mock server ws",
		RunE: func(cmd *cobra.Command, args []string) error {
			server := ServerDemo{}

			server.Start("server_1", opts.protocol, opts.addr)

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol tcp or ws")
	cmd.PersistentFlags().StringVarP(&opts.addr, "addr", "a", ":9001", "server address")

	return cmd
}
