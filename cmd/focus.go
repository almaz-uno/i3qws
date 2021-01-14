/*
Copyright © 2021 Maxim Kovrov

*/
package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const focusURL = "http://localhost/focus/"

// focusCmd represents the focus command
var focusCmd = &cobra.Command{
	Use:   "focus",
	Short: "Focus specified window in saved history",
	Long: `Focus window by number in the window queue. Number starts from 0 (current focused window), 1 - previous focused window and so on.

Negative number mean posion from the tail of the queue: -1 - the last window, -2 - one from the tail and so on.`,
	Run: func(cmd *cobra.Command, args []string) {
		// context should be canceled while Int signal will be caught
		ctx, cancel := context.WithCancel(context.Background())

		// main processing loop
		retChan := make(chan error, 1)
		go func() {
			err2 := doFocus(ctx, args[0])
			if err2 != nil {
				retChan <- err2
			}
			close(retChan)
		}()

		// Listening OS signals
		quit := make(chan os.Signal, 1)
		go func() {
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			logrus.Warnf("Signal '%s' was caught. Exiting", <-quit)
			cancel()
		}()

		// Listening for the main loop response
		if e := <-retChan; e != nil {
			logrus.WithError(e).Info("Exiting.")
		} else {
			logrus.Info("Exiting.") // it seems to be an nonexistent exodus
		}
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(focusCmd)
}

func doFocus(ctx context.Context, numStr string) error {
	socket := viper.GetString(socketFileSett)
	if len(socket) == 0 {
		return fmt.Errorf("%w: %s", errSettingUnspecified, socketFileSett)
	}

	cl := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", socket)
			},
		},
	}

	url := focusURL + numStr
	resp, err := cl.Get(url)
	if resp != nil {
		defer resp.Body.Close() // nolint: errcheck
	}

	if err != nil {
		return fmt.Errorf("unable to get %s: %w", url, err)
	}

	bb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithError(err).Debugf("Unable to read answer from %s", url)
	}
	logrus.WithField("answer", string(bb)).Debugf("Successfully changed to window number %s", numStr)

	return err
}
