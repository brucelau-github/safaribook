package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmaupin/go-epub"
	"github.com/spf13/cobra"
)

var cookieFile string
var wait int
var epubCmd = &cobra.Command{
	Use:   "epub command",
	Short: "convert html to epub",
	Long:  "convert html to epub",
	Run:   epubRun,
}

func init() {
	epubCmd.Flags().StringVarP(&cookieFile, "cookie", "k", "~/.safaricookie", "oreilly website cookie, read from ~/.safaricookie by default")
	epubCmd.Flags().IntVarP(&wait, "wait", "w", -1, "sleep time between each request")
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

func (cli *oreillyClient) doGetFile(url string, header *http.Header, w io.Writer) {
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
	_, err = io.Copy(w, resp.Body)
	cobra.CheckErr(err)
}

func (cli *oreillyClient) doGetImage(imgUrl string, header *http.Header) string {
	link, _ := url.Parse(imgUrl)
	_, fileName := filepath.Split(link.Path)
	body := bytes.Buffer{}
	cli.doGetFile(imgUrl, header, &body)
	targetPath := filepath.Join(os.TempDir(), fileName)
	f, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE, 0755)
	cobra.CheckErr(err)
	io.Copy(f, &body)
	return targetPath
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

func epubRun1(cmd *cobra.Command, args []string) {
	cookie, err := ioutil.ReadFile(cookieFile)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("Fail to open cookie file %s, %q", cookieFile, err))
	}

	header := &http.Header{
		// "Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
		// "Accept-Language": {"en-US,en;q=0.5"},
		// "Connection":      {"keep-alive"},
		// "Host":            {"learning.oreilly.com"},
		// "User-Agent":      {"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:59.0) Gecko/20100101 Firefox/59.0"},
		// "Cookie":          {string(bytes.TrimSpace(cookie))},
	}
	header.Set("Cookie", string(bytes.TrimSpace(cookie)))
	webCli := newOreillyCilent()
	req, err := http.NewRequest("GET", "https://learning.oreilly.com/api/v2/epubs/urn:orm:book:9780134686097/files/ch1.xhtml", nil)
	// for _, line := range strings.Split(string(bytes.TrimSpace(cookie)), ";") {
	// 	sect := strings.Split(strings.TrimSpace(line), "=")
	// 	req.AddCookie(&http.Cookie{Name: sect[0], Value: sect[1]})
	// }
	req.Header.Set("Cookie", string(bytes.TrimSpace(cookie)))
	cobra.CheckErr(err)
	// req.Header = *header
	resp, err := webCli.Do(req)
	fmt.Println(len(req.Header.Get("Cookie")))
	cobra.CheckErr(err)
	defer resp.Body.Close()
	// if resp.StatusCode != 200 {
	msg := bytes.Buffer{}
	msg.ReadFrom(resp.Body)
	fmt.Println(msg.String())
	// 	cobra.CheckErr(fmt.Errorf("Request Failed %s: status %d, body %q", url, resp.StatusCode, msg.String()))
	// }
	// dec := json.NewDecoder(resp.Body)
	// err = dec.Decode(target)
	// cobra.CheckErr(err)
}

func epubRun2(cmd *cobra.Command, args []string) {
	bk := struct {
		Chapters     string
		Title        string
		ISBN         string
		Descriptions struct {
			Text string `json:"text/plain"`
		}
	}{}
	data, err := ioutil.ReadFile("tmp2.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &bk); err != nil {
		panic("unmarshal code.json: " + err.Error())
	}
	fmt.Fprintln(cmd.OutOrStdout(), bk)
}
func epubRun(cmd *cobra.Command, args []string) {
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
	bk := struct {
		Chapters     string
		Title        string
		ISBN         string
		Descriptions struct {
			Text string `json:"text/plain"`
		}
	}{}
	webCli.doGet("https://learning.oreilly.com/api/v2/epubs/urn:orm:book:"+bookID, header, &bk)
	fmt.Fprintf(cmd.OutOrStdout(), "found a book Title: %q\n", bk.Title)
	chapters := struct {
		Count   int
		Results []struct {
			Title         string
			ContentURL    string `json:"content_url"`
			RelatedAssets struct {
				Images []string
			} `json:"related_assets"`
		}
	}{}
	webCli.doGet(bk.Chapters, header, &chapters)
	fmt.Fprintf(cmd.OutOrStdout(), "total chapters: %d\n", len(chapters.Results))
	ebk := epub.NewEpub(bk.Title)
	ebk.SetDescription(bk.Descriptions.Text)
	if err != nil {
		log.Fatal(err)
	}

	imageCaches := []string{}
	for i, cht := range chapters.Results {
		if wait > 0 {
			time.Sleep(time.Duration(wait) * time.Second)
		}
		buf := bytes.Buffer{}
		webCli.doGetFile(cht.ContentURL, header, &buf)
		body := buf.String()
		for _, imgUrl := range cht.RelatedAssets.Images {
			imgPath := webCli.doGetImage(imgUrl, header)
			imageCaches = append(imageCaches, imgPath)
			imgPathEpub, _ := ebk.AddImage(imgPath, "")
			rlink, _ := url.Parse(imgUrl)
			body = strings.ReplaceAll(body, rlink.Path, imgPathEpub)
			if i == 0 {
				ebk.SetCover(imgPathEpub, "")
			}
		}
		ebk.AddSection(body, cht.Title, "", "")
	}
	err = ebk.Write(bookID + ".epub")
	cobra.CheckErr(err)
	for _, img := range imageCaches {
		os.Remove(img)
	}
}
