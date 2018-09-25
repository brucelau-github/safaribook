package main

import (
	"bytes"
	"strings"
	"testing"
)

//func TestHtmlParse(t *testing.T) {
//	file, err := os.Open("ch01.html") // For read access.
//	if err != nil {
//		t.Fatal(err)
//	}
//	defer file.Close()
//	doc, err := html.Parse(file)
//	if err != nil {
//		t.Fatal("error when parsing html page", err)
//	}
//	var chp *html.Node
//	var f func(*html.Node)
//	f = func(n *html.Node) {
//		if n.Type == html.ElementNode && n.Data == "div" {
//			for _, a := range n.Attr {
//				if a.Key == "id" && a.Val == "sbo-rt-content" {
//					chp = n
//					break
//				}
//			}
//		}
//		for c := n.FirstChild; c != nil; c = c.NextSibling {
//			f(c)
//		}
//	}
//	f(doc)
//	var printNode func(*html.Node)
//	printNode = func(n *html.Node) {
//		t.Log(n.Data)
//		for c := n.FirstChild; c != nil; c = c.NextSibling {
//			f(c)
//		}
//	}
//	if chp != nil {
//		printNode(doc)
//	}
//
//	//r := bufio.NewReader(f)
//	//content, _ := ioutil.ReadAll(file)
//	//t.Log(string(content))
//	//	data := bufio.NewReader(file)
//
//	//z := html.NewTokenizer(strings.NewReader("<div>this is a paragraph</div>"))
//	//	z := html.NewTokenizer(r)
//	//loop:
//	//	for {
//	//		tt := z.Next()
//	//		switch tt {
//	//		case html.ErrorToken:
//	//			t.Logf("%v", z.Err())
//	//			break loop
//	//		//case html.TextToken:
//	//		//	t.Log(z.Text())
//	//		case html.StartTagToken, html.EndTagToken:
//	//			tkn := z.Token()
//	//			if strings.Contains(tkn.String(), `id="sbo-rt-content"`) {
//	//				t.Log(string(z.Raw()))
//	//			}
//	//			//t.Log(tkn.String())
//	//			//tn, moreAttr := z.TagName()
//	//			//if bytes.Equal(tn, []byte("div") && moreAttr {
//	//			//	for moreAttr {
//	//			//		key, val, moreAttr = z.Attr()
//	//			//		if bytes.Equal(key, []byte("id")) &&
//	//			//			bytes.Equal(val, []byte("sbo-rt-content")) {
//	//
//	//			//			t.Log(z.Token().String())
//	//			//		}
//	//			//	}
//	//			//}
//	//		}
//	//	}
//}
//
//func TestParse(t *testing.T) {
//	_, err := os.Open(imageDirecotry)
//	if os.IsNotExist(err) {
//		if err := os.Mkdir(imageDirecotry, 0755); err != nil {
//			log.Print("fail to creat a file")
//		}
//		log.Print("create new directory", imageDirecotry)
//	} else {
//		log.Print("using existing directory", imageDirecotry)
//	}
//
//	f, _ := os.Open("Linux_Kernel_Development-_Third_Edition.html")
//	defer f.Close()
//	data, _ := ioutil.ReadAll(f)
//	t.Log(parseImages(string(data)))
//	//images := parseImages(string(data))
//	//for _, img := range images {
//	//	log.Print("downloading image: ", img)
//	//	downloadImage(img)
//	//}
//}

func TestCleanURL(t *testing.T) {
	testcases := []struct {
		html, expected string
	}{
		{`<a href="ix01.html#idm140582977822192">Index</a>`, `<a href="#idm140582977822192">Index</a>`},
		{`<img src="assets/cover.png">`, `<img src="cover.png">`},
	}
	for _, tt := range testcases {
		buf := bytes.NewBufferString(tt.html)
		cleanHTML(buf)
		if buf.String() != tt.expected {
			t.Fatalf("fail cleanURL got:%s, expected: %s", buf.String(), tt.expected)
		}
	}

}

func TestParseHtml(t *testing.T) {
	var chcontent = `
<div id="sbo-rt-content"><section>


            <article>



            </article>


        </section>
    </div><div id="sbo-rt-content"><section>

                            <header>
                    <h1 class="header-title">Mastering Go</h1>
                </header>

            <article>

<p style="font-size: 11px">All rights reserved. No part of this book may be reproduced, stored in a retrieval system, or transmitted in any form or by any means, without the prior written permission of the publisher, except in the case of brief quotations embedded in critical articles or reviews.</p>
<p style="font-size: 11px">Every effort has been made in the preparation of this book to ensure the accuracy of the information presented. However, the information contained in this book is sold without warranty, either express or implied. Neither the author, nor Packt Publishing or its dealers and distributors, will be held liable for any damages caused or alleged to have been caused directly or indirectly by this book.</p>
<p style="font-size: 11px">Packt Publishing has endeavored to provide trademark information about all of the companies and products mentioned in this book by the appropriate use of capitals. However, Packt Publishing cannot guarantee the accuracy of this information.</p>
<p style="font-size: 11px"><strong>Acquisition Editors</strong>: <span>Frank Pohlmann, Suresh Jain</span><br>
<strong>Project Editor</strong>: <span>Kishor Rit</span><br>
<strong>Graphics</strong>: Tom Scaria<br>
<strong>Production Coordinator</strong>: Shantanu Zagade</p>
<p style="font-size: 11px">First published: April 2018</p>
<p style="font-size: 11px">Published by Packt Publishing Ltd.<br>
Livery Place<br>
35 Livery Street<br>
Birmingham<br>
B3 2PB, UK.</p>
<p style="font-size: 11px">ISBN <span class="sugar_field">978-1-78862-654-5</span></p>
<p style="font-size: 11px"><a href="http://www.packtpub.com" target="_blank">www.packtpub.com</a></p>


            </article>


        </section>
    </div>`
	parseHTML(strings.NewReader(chcontent))
	t.Log("pass")
}
