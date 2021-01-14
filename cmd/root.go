/*
Copyright © 2021 Maxim Kovrov
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var errSettingUnspecified = errors.New("setting unspecified")

const (
	logLevelSett   = "log.level"
	socketFileSett = "socket-file"

	socketFileDefault = ".i3qws.sock"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "i3qws",
	Short: "Quick select windows for i3wm.",
	Long: `Sometimes it's good idea — switch window, not only workspaces in i3wm.
	
	And i3qws will bring this ability to our favorite window manager.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if level, e := logrus.ParseLevel(viper.GetString(logLevelSett)); e == nil {
			logrus.SetLevel(level)
		} else {
			logrus.WithError(e).Warnf("Unable to parse log level '%s'", viper.GetString(logLevelSett))
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.i3qws.yaml)")
	rootCmd.PersistentFlags().StringP(socketFileSett, "s", "", "socket file for communication (default is $HOME/"+socketFileDefault+")")
	rootCmd.PersistentFlags().StringP(logLevelSett, "L", "info", "logrus logging level")

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		panic("unable to bind flags " + err.Error())
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".i3qws" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".i3qws")

		viper.SetDefault(socketFileSett, filepath.Join(home, socketFileDefault))

	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
