package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage skills-cli configuration",
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val := viper.Get(args[0])
		if val == nil {
			return fmt.Errorf("key %q not found", args[0])
		}
		fmt.Println(val)
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.Set(args[0], args[1])
		cfgFile := viper.ConfigFileUsed()
		if cfgFile == "" {
			return fmt.Errorf("no config file in use; create ~/.skills-cli/config.yaml first")
		}
		return viper.WriteConfig()
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all config values",
	Run: func(cmd *cobra.Command, args []string) {
		for k, v := range viper.AllSettings() {
			fmt.Printf("%s = %v\n", k, v)
		}
	},
}

func init() {
	configCmd.AddCommand(configGetCmd, configSetCmd, configListCmd)
	rootCmd.AddCommand(configCmd)
}
