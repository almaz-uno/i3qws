/*
Copyright © 2021 Maxim Kovrov

*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cured-plumbum/i3qws/pkg/i3qws"
	"github.com/cured-plumbum/i3qws/pkg/serve"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/net"
)

const (
	shutdownTimeout = 10 * time.Second

	markFormatSett = "mark-format"
	restartSett    = "restart"
)

var errAnotherInstanceRunning = errors.New("another instance is started")

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run",
	Aliases: []string{"start"},
	Short:   "Runs application for listening i3wm window events and log windows in the log",
	Long: `i3qws subscribes to windows change events and remembers in the memory windows got focus.
User can bring up any window with 'focus' command.

Warning! Windows list always clears in case of restart i3wm.`,
	Run: func(cmd *cobra.Command, args []string) {
		doMain(doRun)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringP(markFormatSett, "m", "%d", "marks format; not marks used, if is empty")
	runCmd.PersistentFlags().BoolP(restartSett, "R", false, "force to restart service if it already started")

	if err := viper.BindPFlags(runCmd.PersistentFlags()); err != nil {
		panic("unable to bind flags " + err.Error())
	}
}

func doRun(ctx context.Context) error {
	socket := viper.GetString(socketFileSett)
	if len(socket) == 0 {
		return fmt.Errorf("%w: %s", errSettingUnspecified, socketFileSett)
	}

	_, err := getURL(ctx, listURL)
	if err == nil {
		if !viper.GetBool(restartSett) {
			return fmt.Errorf("%w on %s", errAnotherInstanceRunning, socket)
		}
		_, e2 := getURL(ctx, stopURL)
		if e2 != nil {
			logrus.WithError(e2).Warn("Unable to stop existing instance")
		}

		time.Sleep(time.Second)

	} else if net.IsConnectionRefused(err) {
		os.Remove(socket)
	}

	logrus.Info("Successfully starting main loop")

	stopCh := make(chan bool)

	i3qws := i3qws.DoSpy(ctx, viper.GetString(markFormatSett))
	echo, err := serve.EchoServe(ctx, stopCh, i3qws, socket)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
	case <-stopCh:
		logrus.Info("Request to stop is got. Stopping.")
	}

	sctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err = echo.Shutdown(sctx)
	logrus.Info("Successfully exiting main loop")
	return err
}
