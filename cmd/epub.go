package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

type book struct {
	Chapters string
	Title    string
	ISBN     string
	id       string
}

var cookieFile string
var epubCmd = &cobra.Command{
	Use:   "epub command",
	Short: "convert html to epub",
	Long:  "convert html to epub",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cobra.CheckErr(fmt.Errorf("epub needs a bookid for the command"))
		}
		bookID := args[0]

		cookie, err := ioutil.ReadFile(cookieFile)
		if err != nil {
			cobra.CheckErr(fmt.Errorf("Fail to open cookie file %s, %q", cookieFile, err))
		}

		header := &http.Header{
			"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
			"Accept-Language": {"en-US,en;q=0.5"},
			"Connection":      {"keep-alive"},
			"Host":            {"learning.oreilly.com"},
			"User-Agent":      {"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:59.0) Gecko/20100101 Firefox/59.0"},
			"Cookie":          {string(bytes.TrimSpace(cookie))},
		}
		webCli := newOreillyCilent()
		bk := book{}
		webCli.doGet("https://learning.oreilly.com/api/v2/epubs/urn:orm:book:"+bookID, header, &bk)
		fmt.Fprintln(cmd.OutOrStdout(), bk)
	},
}

func init() {
	epubCmd.Flags().StringVarP(&cookieFile, "cookie", "k", "~/.safaricookie", "oreilly website cookie, read from ~/.safaricookie by default")
}

type oreillyClient struct {
	http.Client
}

func (cli *oreillyClient) doGet(url string, header *http.Header, target interface{}) {
	req, err := http.NewRequest("GET", url, nil)
	cobra.CheckErr(err)
	req.Header = *header
	resp, err := cli.Do(req)
	cobra.CheckErr(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		msg := bytes.Buffer{}
		msg.ReadFrom(resp.Body)
		cobra.CheckErr(fmt.Errorf("Request Failed %s: status %d, body %q", url, resp.StatusCode, msg.String()))
	}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(target)
	cobra.CheckErr(err)
}

func newOreillyCilent() *oreillyClient {
	return &oreillyClient{
		Client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: true,
			},
		},
	}
}
