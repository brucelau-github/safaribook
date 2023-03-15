package cmd

import (
	"fmt"
	"os"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
)

var rawText, outPath string
var size int
var qrcodeCmd = &cobra.Command{
	Use:   "qrcode command",
	Short: "qr code generator",
	Long:  "qr code generator",
	Run:   qrcodeRun,
}

func init() {
	qrcodeCmd.Flags().StringVarP(&rawText, "text", "t", "", "text to be encoded")
	qrcodeCmd.Flags().StringVarP(&outPath, "out", "o", "qr.png", "file name of the qrfile")
	qrcodeCmd.Flags().IntVarP(&size, "size", "s", 256, "QR image size")
	qrcodeCmd.MarkFlagRequired("text")
}

func qrcodeRun(cmd *cobra.Command, args []string) {
	err := qrcode.WriteFile(rawText, qrcode.Medium, size, outPath)
	cobra.CheckErr(err)
	fmt.Fprintf(os.Stdout, "qrcode file generated: %s", rawText)
}
