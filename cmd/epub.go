package cmd

import (
	"bytes"
	"compress/gzip"
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
var wait, retry int
var epubCmd = &cobra.Command{
	Use:   "epub command",
	Short: "convert html to epub",
	Long:  "convert html to epub",
	Run:   epubRun,
}

type readerAndCloser struct {
	io.Reader
	io.Closer
}

func init() {
	epubCmd.Flags().StringVarP(&cookieFile, "cookie", "k", "~/.zeenrc", "oreilly website cookie, read from ~/.zeenrc by default")
	epubCmd.Flags().IntVarP(&wait, "wait", "w", -1, "sleep time between each request")
	epubCmd.Flags().IntVarP(&retry, "retry", "t", 3, "times of retry before abort the connection")
}

type oreillyClient struct {
	http.Client
	timesOfAttempt int
}

func (cli *oreillyClient) doGet(url string, header *http.Header, target interface{}) {
	req, err := http.NewRequest("GET", url, nil)
	cobra.CheckErr(err)
	req.Header = *header
	resp := cli.doRequestWithAttempts(req)
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(target)
	cobra.CheckErr(err)
}

func (cli *oreillyClient) doGetFile(url string, header *http.Header, w io.Writer) {
	req, err := http.NewRequest("GET", url, nil)
	cobra.CheckErr(err)
	req.Header = *header
	resp := cli.doRequestWithAttempts(req)
	cobra.CheckErr(err)
	defer resp.Body.Close()
	_, err = io.Copy(w, resp.Body)
	cobra.CheckErr(err)
}

func (cli *oreillyClient) doGetImage(imgURL string, header *http.Header) string {
	link, _ := url.Parse(imgURL)
	_, fileName := filepath.Split(link.Path)
	body := bytes.Buffer{}
	cli.doGetFile(imgURL, header, &body)
	targetPath := filepath.Join(os.TempDir(), fileName)
	f, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE, 0755)
	cobra.CheckErr(err)
	io.Copy(f, &body)
	return targetPath
}

func (cli *oreillyClient) login(header *http.Header) {
	req, err := http.NewRequest("GET", "https://api.oreilly.com/api/v2/me", nil)
	cobra.CheckErr(err)
	req.Header = *header
	resp := cli.doRequestWithAttempts(req)
	defer resp.Body.Close()
}

func (cli *oreillyClient) doRequestWithAttempts(req *http.Request) *http.Response {
	var resp *http.Response
	var err error
	for i := 0; i < cli.timesOfAttempt; i++ {
		resp, err = cli.Do(req)
		if err == nil && resp.StatusCode == 200 {
			if h := resp.Header.Get("ETag"); strings.HasSuffix(h, "-sample") {
				cobra.CheckErr(fmt.Errorf("Sample Response - Request Failed %s", req.URL.String()))
				return nil
			}
			if h := resp.Header.Get("Content-Encoding"); strings.Contains(h, "gzip") {
				gzReader, err := gzip.NewReader(resp.Body)
				cobra.CheckErr(err)
				resp.Body = &readerAndCloser{gzReader, resp.Body}
			}
			return resp
		}
		time.Sleep(time.Duration(i) * time.Second)
		fmt.Fprintf(os.Stdout, "attempt to get %s; failed: %d/%d\n", req.URL.String(), i+1, cli.timesOfAttempt)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error %q\n", err)
		} else {
			msg := bytes.Buffer{}
			msg.ReadFrom(resp.Body)
			fmt.Fprintf(os.Stdout, "detail response body: %q\n", msg.String())
		}
	}
	cobra.CheckErr(fmt.Errorf("Request Failed %s", req.URL.String()))
	return nil
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
		timesOfAttempt: 3,
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
	resp := webCli.doRequestWithAttempts(req)
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

func epubRun(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cobra.CheckErr(fmt.Errorf("epub needs a bookid for the command"))
	}
	bookID := args[0]
	if strings.HasPrefix(cookieFile, "~/") {
		homeDir, _ := os.UserHomeDir()
		cookieFile = filepath.Join(homeDir, cookieFile[2:])
	}

	cookie, err := ioutil.ReadFile(cookieFile)
	if err != nil {
		cobra.CheckErr(fmt.Errorf("Fail to open cookie file %s, %q", cookieFile, err))
	}

	header := &http.Header{}

	buf := bytes.NewBuffer(cookie)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		hdr := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(hdr) != 2 {
			continue
		}
		header.Set(strings.TrimSpace(hdr[0]), strings.TrimSpace(hdr[1]))
	}

	webCli := newOreillyCilent()
	if retry >= 1 {
		webCli.timesOfAttempt = retry
	}

	if bear := header.Get("Authorization"); bear != "" {
		webCli.login(header)
		header.Del("Authorization")
	}

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
	type chapter struct {
		Title         string
		ContentURL    string `json:"content_url"`
		RelatedAssets struct {
			Images []string
		} `json:"related_assets"`
	}
	chapters := []chapter{}

	type chapterIndex struct {
		Count   int
		Next    string
		Results []chapter
	}
	chptIndex := chapterIndex{}
	webCli.doGet(bk.Chapters, header, &chptIndex)
	chapters = append(chapters, chptIndex.Results...)
	for chptIndex.Next != "" {
		nextIndex := chapterIndex{}
		webCli.doGet(chptIndex.Next, header, &nextIndex)
		chapters = append(chapters, nextIndex.Results...)
		chptIndex = nextIndex
	}

	fmt.Fprintf(cmd.OutOrStdout(), "total chapters: %d\n", len(chapters))
	ebk := epub.NewEpub(bk.Title)
	ebk.SetDescription(bk.Descriptions.Text)
	if err != nil {
		log.Fatal(err)
	}

	imageCaches := []string{}
	totalChapters := len(chapters)
	for i, cht := range chapters {
		fmt.Fprintf(cmd.OutOrStdout(), "process chapter %d/%d: %q -> %q\n", i+1, totalChapters, cht.Title, cht.ContentURL)
		if wait > 0 {
			time.Sleep(time.Duration(wait) * time.Second)
		}
		buf := bytes.Buffer{}
		webCli.doGetFile(cht.ContentURL, header, &buf)
		body := buf.String()
		for _, imgURL := range cht.RelatedAssets.Images {
			imgPath := webCli.doGetImage(imgURL, header)
			imageCaches = append(imageCaches, imgPath)
			imgPathEpub, _ := ebk.AddImage(imgPath, "")
			rlink, _ := url.Parse(imgURL)
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
