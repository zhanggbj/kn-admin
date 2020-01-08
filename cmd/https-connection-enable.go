/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"io/ioutil"
	"istio.io/api/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	istioClientSet "istio.io/client-go/pkg/clientset/versioned"
)

type cmdFlags struct {
	TLSCertFile string
	// TLSKeyFile is the path to a TLS key file
	TLSKeyFile string
}

var flags cmdFlags

// enableCmd represents the enable command
var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		kubeConfig := os.Getenv("KUBECONFIG")
		if kubeConfig == "" {
			fmt.Println("cannot get cluster kube config, please export environment variable KUBECONFIG")
			os.Exit(1)
		}

		cert, err := ioutil.ReadFile(flags.TLSCertFile)
		if err != nil {
			fmt.Println("cannot read tls cert file")
			os.Exit(1)
		}

		key, err := ioutil.ReadFile(flags.TLSKeyFile)
		if err != nil {
			fmt.Println("cannot read tls key file")
			os.Exit(1)
		}

		secretData := map[string][]byte{
			"tls.crt": cert,
			"tls.key": key,
		}

		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			Type: corev1.SecretTypeTLS,
			ObjectMeta: metav1.ObjectMeta{
				Name:      "istio-ingressgateway-certs",
				Namespace: "istio-system",
			},
			Data: secretData,
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

		_, err = clientset.CoreV1().Secrets("istio-system").Get("istio-ingressgateway-certs", metav1.GetOptions{})
		if err == nil {
			err = clientset.CoreV1().Secrets("istio-system").Delete("istio-ingressgateway-certs", &metav1.DeleteOptions{})
			if err != nil {
				fmt.Println("failed to delete secret:", err)
				os.Exit(1)
			}
		}

		_, err = clientset.CoreV1().Secrets("istio-system").Create(secret)
		if err != nil {
			fmt.Println("failed to create secret:", err)
			os.Exit(1)
		}

		ic, err := istioClientSet.NewForConfig(cfg)
		if err != nil {
			fmt.Println("Failed to create istio client: %s", err)
			os.Exit(1)
		}

		istioGateway, err := ic.NetworkingV1alpha3().Gateways("knative-serving").Get("knative-ingress-gateway", metav1.GetOptions{})
		if err != nil {
			fmt.Println("Failed to get VirtualService in %s namespace: %s", "knative-serving", err)
			os.Exit(1)
		}

		host := []string{"*"}
		httpsServer := v1alpha3.Server{
			Hosts: host,
			Port: &v1alpha3.Port{
				Name:     fmt.Sprintf("%s", "https"),
				Number:   443,
				Protocol: "https",
			},
			Tls: &v1alpha3.Server_TLSOptions{
				Mode:              v1alpha3.Server_TLSOptions_SIMPLE,
				PrivateKey:        "/etc/istio/ingressgateway-certs/tls.key",
				ServerCertificate: "/etc/istio/ingressgateway-certs/tls.crt",
			},
		}

		desiredIstioGw := istioGateway.DeepCopy()
		servers := istioGateway.Spec.Servers
		desiredIstioGw.Spec.Servers = append(servers, &httpsServer)

		_, err = ic.NetworkingV1alpha3().Gateways("knative-serving").Update(desiredIstioGw)
		if err != nil {
			fmt.Printf("could not update Gateway '%s/%s':", "knative-serving", "knative-ingress-gateway", err)
			os.Exit(1)
		}

		fmt.Printf("Enabled feature flag https-connection\n")

	},
}


func init() {
	httpsConnectionCmd.AddCommand(enableCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// enableCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// enableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	enableCmd.Flags().StringVar(&flags.TLSCertFile, "tls-cert", "", "Path to TLS certificate file")
	enableCmd.Flags().StringVar(&flags.TLSKeyFile, "tls-key", "", "Path to TLS key file")
}
