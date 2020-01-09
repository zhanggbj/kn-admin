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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	"github.com/spf13/cobra"
	"encoding/json"
)

type prcmdFlags struct {
	DockerServer string
	SecretName string
	DockerEmail string
	DockerUsername string
	DockerPassword string
}

type DockerRegistry struct {
	Auths Auths `json:"auths"`
}
type UsIcrIo struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
	Email    string `json:"Email"`
}
type Auths struct {
	UsIcrIo UsIcrIo `json:"us.icr.io"`
}

var prflags prcmdFlags

// enableCmd represents the enable command
var prEnableCmd = &cobra.Command{
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

		dockerCfg := DockerRegistry{
			Auths: Auths{
				UsIcrIo: UsIcrIo{
					Username: prflags.DockerUsername,
					Password: prflags.DockerPassword,
					Email: prflags.DockerEmail,
				},
			},
		}

		j, err := json.Marshal(dockerCfg)
		if err != nil {
			panic(err)
		}

		secretData := map[string][]byte{
			".dockerconfigjson": j,
		}

		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			Type: corev1.SecretTypeDockerConfigJson,
			ObjectMeta: metav1.ObjectMeta{
				Name:      prflags.SecretName,
				Namespace: "default",
			},
			Data: secretData,
		}

		_, err = clientset.CoreV1().Secrets("default").Create(secret)
		if err != nil {
			fmt.Println("failed to create secret:", err)
			os.Exit(1)
		}

		defaultSa, err := clientset.CoreV1().ServiceAccounts("default").Get("default", metav1.GetOptions{})
		desiredSa := defaultSa.DeepCopy()
		desiredSa.ImagePullSecrets = []corev1.LocalObjectReference{{
				Name: prflags.SecretName,
			},}

		_, err = clientset.CoreV1().ServiceAccounts("default").Update(desiredSa)
		if err != nil {
			fmt.Println("failed to add registry secret in default Service Account:", err)
			os.Exit(1)
		}

		fmt.Printf("Private registry %s enabled for default Service Account", prflags.DockerServer)
	},
}

func init() {
	privateRegistryCmd.AddCommand(prEnableCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// enableCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// enableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	prEnableCmd.Flags().StringVar(&prflags.SecretName, "secret-name", "", "Registry Secret Name")
	prEnableCmd.Flags().StringVar(&prflags.DockerServer, "docker-server", "", "Registry Address")
	prEnableCmd.Flags().StringVar(&prflags.DockerEmail, "docker-email", "", "Registry Email")
	prEnableCmd.Flags().StringVar(&prflags.DockerUsername, "docker-username", "", "Registry Username")
	prEnableCmd.Flags().StringVar(&prflags.DockerPassword, "docker-password", "", "Registry Email")
}
