package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	patchAPIServer     bool
	genKubeConfigCerts bool

	// installCmd represents the install command
	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install Konnectivity Server and Agents.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("to be implemented")
		},
	}
)

func init() {
	konnectivityCmd.AddCommand(installCmd)
	installCmd.PersistentFlags().BoolVarP(&patchAPIServer, "patch-apiserver", "p", false, "Patch Kube APIServer")
	installCmd.PersistentFlags().BoolVarP(&patchAPIServer, "gen-certs", "g", false, "Generate Kubeconfig certificates for Konnectivity.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
