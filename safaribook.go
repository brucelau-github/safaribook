package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var (
	bookAPI      = "https://www.safaribooksonline.com/api/v1/book/"
	httpClient   *http.Client
	commonHeader *http.Header
	htmlHeader   = ` <!DOCTYPE html>
<html>
	<head>
	<meta charset="UTF-8">
	<title>%s</title>
	<link rel="stylesheet" href="stylesheet.css">
	</head>
	<body>
`
	htmlTail = `
	</body>
</html>
`
	bookContent, cssContent bytes.Buffer
	downloadedFiles         = map[string]bool{}
)

type book struct {
	Chapters    []string `json:"chapters"`
	Cover       string   `json:"cover"`
	ChapterList string   `json:"chapterlist"`
	Title       string   `json:"title"`
	ID          string   `json:"identifier"`
}

type chapter struct {
	ContentURL   string `json:"content"`
	Stylesheets  []stylesheet
	Images       []string `json:"images"`
	AssetBaseURL string   `json:"asset_base_url"`
	FullPath     string   `json:"full_path"`
}

type stylesheet struct {
	URL      string `json:"url"`
	FullPath string `json:"full_path"`
}

type css struct {
	Content string `json:"content"`
}

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}
	commonHeader = &http.Header{
		"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
		"Accept-Language": {"en-US,en;q=0.5"},
		"Connection":      {"keep-alive"},
		"Host":            {"www.safaribooksonline.com"},
		"User-Agent":      {"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:59.0) Gecko/20100101 Firefox/59.0"},
	}
}

func retrivesBookInfo(bookID string) (bk book) {
	body := doGet(bookAPI + bookID)
	dec := json.NewDecoder(&body)
	if err := dec.Decode(&bk); err != nil {
		log.Fatal("fail to decode book json", err)
	}
	return
}

func downladBook(bookID string) {
	log.Println("retriveing book page")
	bk := retrivesBookInfo(bookID)
	log.Printf("got book : %s, with %d chapters.", bk.Title, len(bk.Chapters))

	bookContent.WriteString(fmt.Sprintf(htmlHeader, bk.Title))

	for _, chURI := range bk.Chapters {
		log.Println("retriving chapter", chURI)
		ch := getChapter(chURI)
		body := doGet(ch.ContentURL)
		cleanHTML(&body)
		bookContent.ReadFrom(&body)
	}
	bookContent.WriteString(htmlTail)
	saveFile(bookID+".html", &bookContent)
	saveFile("stylesheet.css", &cssContent)
}

func doRequest(method, url string) (buf bytes.Buffer) {
	req, _ := http.NewRequest(method, url, nil)
	req.Header = *commonHeader
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("fail to request %s, error: %v\n", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("wrong response status code: ", resp.StatusCode)
	}

	buf.ReadFrom(resp.Body)
	return
}

func doGet(url string) (buf bytes.Buffer) {
	return doRequest("GET", url)
}

func getChapter(chURI string) (ch chapter) {
	body := doGet(chURI)

	dec := json.NewDecoder(&body)
	if err := dec.Decode(&ch); err != nil {
		log.Fatal("fail to decode chapter: ", err)
	}

	for _, img := range ch.Images {
		downloadimg(img, ch.AssetBaseURL+img)
	}
	for _, css := range ch.Stylesheets {
		r := downloadFile(css.URL)
		if r != nil {
			cssContent.ReadFrom(r)
		}
	}
	return ch

}

func downloadimg(img, url string) {
	dir, _ := filepath.Split(img)
	if dir != "" {
		_, err := os.Stat(dir)
		if err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0755); err != nil {
				log.Fatal("fail to create image directory")
			}
		}

	}
	r := downloadFile(url)
	if r != nil {
		saveFile(img, r)
	}
}

func downloadFile(fileURL string) io.Reader {
	body := doGet(fileURL)
	u, _ := url.Parse(fileURL) // if we run here there are no error in parse url
	_, filename := filepath.Split(u.Path)
	_, hasDownload := downloadedFiles[filename]
	if hasDownload {
		log.Printf("%s has exists.\n", filename)
		return nil
	}
	downloadedFiles[filename] = true
	log.Println("downloading", filename)

	return &body

}

func saveFile(filename string, content io.Reader) {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Printf("%s could not be saved : %v", filename, err)
		return
	}
	defer f.Close()
	io.Copy(f, content)
}

func cleanHTML(buf *bytes.Buffer) {
	content := buf.String()
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, `<div id="sbo-rt-content">`)
	content = strings.TrimSuffix(content, `</div>`)
	buf.Reset()
	subCases := []struct {
		pattern, dst string
	}{
		{`href="[^"]{2,50}(#[^"]{2,100})"`, `href="${1}"`},
	}
	for _, reg := range subCases {

		re, _ := regexp.Compile(reg.pattern)
		content = re.ReplaceAllString(content, reg.dst)
	}
	buf.WriteString(content)

}

func parseHTML(r io.Reader) string {
	var cleanContent bytes.Buffer
	secDepth := 100
	z := html.NewTokenizer(r)
	depth := 0
	for z.Err() == nil {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break
		case html.TextToken:
			if depth > 0 {
				cleanContent.WriteString(cleanText(z.Text()))
			}
		case html.StartTagToken:
			t := z.Token()
			if t.Data != "br" {
				depth++
			}
			if t.Data == "section" {
				secDepth = depth
			}
			fmt.Println(secDepth, depth)
			if depth >= secDepth {
				cleanContent.WriteString(t.String())
			}
		case html.EndTagToken:
			depth--
			if depth >= secDepth {
				cleanContent.WriteString(z.Token().String())
			}
		}
	}
	return cleanContent.String()
}

func cleanText(s []byte) string {
	trimed := bytes.TrimSpace(s)
	return html.EscapeString(string(trimed))
}

func main() {
	var usage = `
usage: safaribook <-c cookiefile> <bookid>

Command Argument:
	bookid	the book id your want to download
`
	cookiefn := flag.String("c", "", "cookie file that authenticate yourself")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if *cookiefn == "" {
		flag.Usage()
		os.Exit(1)
	}
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	cookie, err := ioutil.ReadFile(*cookiefn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading cookie file error", err)
		os.Exit(2)
	}
	commonHeader.Set("Cookie", string(bytes.TrimSpace(cookie)))
	downladBook(flag.Arg(0))
}
