/*
Copyright © 2025 Andrew Melnick meln5674.5674@gmail.com
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/meln5674/vnc-ws-gateway/pkg/gateway"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "vnc-ws-gateway",
	Short:        "Simple remote desktop using VNC over Websockets",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		password := os.Getenv("VNC_WS_GATEWAY_PASSWORD")
		if password == "" {
			return fmt.Errorf("VNC_WS_GATEWAY_PASSWORD is not set")
		}
		passwordFile, err := os.CreateTemp("", "vnc-ws-gateway-passwd-*")
		if err != nil {
			return fmt.Errorf("failed to create password file: %w", err)
		}
		defer os.Remove(passwordFile.Name())
		defer passwordFile.Close()
		mkPasswd := exec.Command("tigervncpasswd", "-f")
		mkPasswd.Stdin = strings.NewReader(password)
		mkPasswd.Stdout = passwordFile
		mkPasswd.Stderr = os.Stderr
		err = mkPasswd.Run()
		if err != nil {
			return fmt.Errorf("failed to create password file: %w", err)
		}
		defer passwordFile.Close()
		slog.Info("listening", "addr", listenAddr)

		srvCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		handlerCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
		defer cancel()

		srv := http.Server{
			Handler: gateway.New(gateway.Config{
				PasswordFile: passwordFile.Name(),
				VNCCmd:       vncCmd,
				VNCArgs:      vncArgs,
				PollPeriod:   vncPollPeriod,
				PollRetries:  vncPollRetries,
			}),
			Addr:        listenAddr,
			BaseContext: func(net.Listener) context.Context { return srvCtx },
		}
		go func() {
			<-srvCtx.Done()
			slog.Info("SIGINT received, shutting down")
			srv.Shutdown(context.Background())
		}()
		go func() {
			<-handlerCtx.Done()
			slog.Info("SIGTERM received, forcibly shutting down")
			srv.Close()
		}()

		return srv.ListenAndServe()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	listenAddr     string
	vncPollPeriod  time.Duration
	vncPollRetries int
	vncCmd         string
	vncArgs        []string
)

func init() {
	rootCmd.Flags().StringVar(&listenAddr, "listen", "127.0.0.1:8080", "Address:port to listen on")
	rootCmd.Flags().DurationVar(&vncPollPeriod, "vnc-poll-period", 100*time.Millisecond, "Time to wait between checks for the VNC server to start listening on its socket")
	rootCmd.Flags().IntVar(&vncPollRetries, "vnc-poll-retries", 10, "Number of retries when dialing VNC server before giving up and reporting and error")
	rootCmd.Flags().StringVar(&vncCmd, "vnc-cmd", "tigervncserver", "VNC server command")
	rootCmd.Flags().StringSliceVar(&vncArgs, "vnc-args", []string{}, "Extra arguments to VNC server")
}
