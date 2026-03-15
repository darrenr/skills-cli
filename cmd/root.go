package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "skills-cli",
	Short: "Browse, install, and manage Agent Skills",
	Long: `skills-cli is a standalone tool for discovering and managing Agent Skills
(SKILL.md files) from public GitHub repositories.

Browse the registry, search by keyword, install skills into your project
or globally, and keep them up to date — no authentication required.

Common workflows:
  skills-cli list
  skills-cli list --installed
  skills-cli search "commit messages"`,
}

// Execute is the entry point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.skills-cli/config.yaml)")
	rootCmd.PersistentFlags().StringP("output", "o", "table", "output format: table, json, yaml")
	if err := viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")); err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not bind output flag:", err)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(filepath.Join(home, ".skills-cli"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("SKILLS")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
