package cmd

import (
	"fmt"
	"os"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var rawText, outPath string
var qrcodeCmd = &cobra.Command{
	Use:   "qrcode command",
	Short: "qr code generator",
	Long:  "qr code generator",
	Run:   qrcodeRun,
}

func init() {
	qrcodeCmd.Flags().StringVarP(&rawText, "text", "t", "", "text to be encoded")
	qrcodeCmd.Flags().StringVarP(&outPath, "out", "o", "qr.png", "file name of the qrfile")
	qrcodeCmd.MarkFlagRequired("text")
}

func qrcodeRun(cmd *cobra.Command, args []string) {
	err := qrcode.WriteFile(rawText, qrcode.Medium, 256, outPath)
	fmt.Fprintf(os.Stdout, "qrcode file generated: %s", rawText)
	cobra.CheckErr(err)
}
