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
	"os"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/tools/clientcmd"
)

var domain string

// updateCmd represents the update command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "set route domain",
	Long: `set Knative route domain for service
For example:
kn admin domain set --custom-domain mydomain.com
`,
	Run: func(cmd *cobra.Command, args []string) {
		kubeConfig := os.Getenv("KUBECONFIG")
		if kubeConfig == "" {
			fmt.Println("cannot get cluster kube config, please export environment variable KUBECONFIG")
			os.Exit(1)
		}

		cfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			fmt.Println("failed to build config:", err)
			os.Exit(1)
		}

		clientset, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			fmt.Println("failed to create client:", err)
			os.Exit(1)
		}

		currentCm := &corev1.ConfigMap{}
		currentCm, err = clientset.CoreV1().ConfigMaps("knative-serving").Get("config-domain", metav1.GetOptions{})
		if err != nil {
			fmt.Println("failed to get ConfigMaps:", err)
			os.Exit(1)
		}

		m := make(map[string]string)
		m[domain] = ""
		desiredCm := currentCm.DeepCopy()
		desiredCm.Data = m

		if !equality.Semantic.DeepEqual(desiredCm.Data, currentCm.Data) {
			_, err = clientset.CoreV1().ConfigMaps("knative-serving").Update(desiredCm)
			if err != nil {
				fmt.Println("failed to update ConfigMaps:", err)
				os.Exit(1)
			}
			fmt.Printf("Updated Knative route domain to %s\n", domain)
		} else {
			fmt.Printf("Knative route domain is already set to %s. Skip update\n", domain)
		}
	},
}

func init() {

	setCmd.Flags().StringVarP(&domain, "custom-domain", "d", "", "Desired custom domain")
	domainCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
