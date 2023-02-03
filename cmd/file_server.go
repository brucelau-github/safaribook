package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var dirPath string
var port int
var fileServerCmd = &cobra.Command{
	Use:   "fsr command",
	Short: "server file system as http server",
	Long:  "open a http server to serve file system",
	Run:   fileServerRun,
}

func init() {
	fileServerCmd.Flags().StringVarP(&dirPath, "dir", "d", ".", "serve directory")
	fileServerCmd.Flags().IntVarP(&port, "port", "p", 8000, "server port")
}

func fileServerRun(cmd *cobra.Command, args []string) {
	http.Handle("/", http.FileServer(http.Dir(dirPath)))

	fmt.Fprintf(os.Stdout, "server path %s is running on http://localhost:%d\nPress Ctrl -C to stop.", dirPath, port)
	cobra.CheckErr(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))

}
