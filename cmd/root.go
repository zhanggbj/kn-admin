/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/zhanggbj/kn-admin/cmd/domain"
	"github.com/zhanggbj/kn-admin/cmd/https-connection"
	"github.com/zhanggbj/kn-admin/cmd/private-registry"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/zhanggbj/kn-admin/gen"
)

var cfgFile string



// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "admin",
	Short: "A plugin of kn client to manage Knative",
	Long: `A plugin of kn client to manage Knative for administrators. 

For example:
kn admin domain set - to set Knative route domain to a custom domain
kn admin https-connection enable - to enable https connection for Knative Service
kn admin private-registry enable - to enable deployment from the private registry
kn admin scale-to-zero enable - to enable scale to zero
kn admin obv profiling get -heap - to get Knative Serving profiling data
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kn-admin.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(domain.NewDomainCmd())
	rootCmd.AddCommand(https_connection.NewHttpsConnectionCmd())
	rootCmd.AddCommand(private_registry.NewPrivateRegistryCmd())
	rootCmd.AddCommand(gen.NewAutoscalingConfigCmd())
	rootCmd.AddCommand(gen.NewAutotlsCmd())
	rootCmd.AddCommand(gen.NewIngressgatewayCmd())
	rootCmd.AddCommand(gen.NewObvCmd())
	rootCmd.AddCommand(gen.NewProfilingCmd())
	rootCmd.AddCommand(gen.NewScaleToZeroCmd())
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

		// Search config in home directory with name ".kn-admin" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kn-admin")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
