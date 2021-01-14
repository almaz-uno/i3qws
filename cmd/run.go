/*
Copyright © 2021 Maxim Kovrov

*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cured-plumbum/i3qws/pkg/i3qws"
	"github.com/cured-plumbum/i3qws/pkg/serve"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const shutdownTimeout = 10 * time.Second

const markFormatSett = "mark-format"

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs application for listening i3wm window events and log windows in the log",
	Long: `i3qws subsribes to windows change events and remembers in the memory windows got focus.
User can bring up any window with 'focus' command.

Please, warning! Windows list will be cleared in case of restart i3wm.`,
	Run: func(cmd *cobra.Command, args []string) {
		// context should be canceled while Int signal will be caught
		ctx, cancel := context.WithCancel(context.Background())

		// main processing loop
		retChan := make(chan error, 1)
		go func() {
			err2 := doRun(ctx)
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
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().StringP(markFormatSett, "m", "%d", "marks format; not marks used, if is empty")

	if err := viper.BindPFlags(runCmd.PersistentFlags()); err != nil {
		panic("unable to bind flags " + err.Error())
	}
}

func doRun(ctx context.Context) error {
	socket := viper.GetString(socketFileSett)
	if len(socket) == 0 {
		return fmt.Errorf("%w: %s", errSettingUnspecified, socketFileSett)
	}

	logrus.Info("Successfully starting main loop")

	i3qws := i3qws.DoSpy(ctx, viper.GetString(markFormatSett))
	echo, err := serve.EchoServe(ctx, i3qws, socket)
	if err != nil {
		return err
	}
	<-ctx.Done()

	sctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err = echo.Shutdown(sctx)
	logrus.Info("Successfully exiting main loop")
	return err
}
