/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type Kubernetes struct {
	Name       string     `json:"name"`
	Version    string     `json:"version"`
	Users      []string   `json:"users,omitempty"`
	NodeNumber func() int `json:"-"`
}

func (k8s Kubernetes) GetUsers() { // method - value receiver
	for _, user := range k8s.Users {
		fmt.Println(user)
	}
}

func (k8s *Kubernetes) AddNewUser(user string) { // method - pointer receiver
	k8s.Users = append(k8s.Users, user)
}

// goBasicCmd represents the goBasic command
var goBasicCmd = &cobra.Command{
	Use:   "go-basic",
	Short: "Run golang basic code",
	Run: func(cmd *cobra.Command, args []string) {
		k8s := Kubernetes{
			Name:    "k8s-demo-cluster",
			Version: "1.31",
			Users:   []string{"alex", "den", "antonio"},
			NodeNumber: func() int {
				return 10
			},
		}
		k8s.GetUsers()
		k8s.AddNewUser("anonymous")
		k8s.GetUsers()

	},
}

func init() {
	rootCmd.AddCommand(goBasicCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// goBasicCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// goBasicCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
