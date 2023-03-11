package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "zeen",
	Short: "zeen toolbox",
	Long:  "zeen toolbox",
}

//Execute the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringP("cookie", "k", "~/.safaricookie", "safari cookie field after login its website")
	rootCmd.AddCommand(epubCmd)
	rootCmd.AddCommand(fileServerCmd)
	rootCmd.AddCommand(qrcodeCmd)
}
