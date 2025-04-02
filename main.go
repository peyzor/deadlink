package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("incorrect number of given arguments expected 1 but got: %d", len(os.Args)-1)
		return
	}

	url := os.Args[1]

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("couldn't get url: %v", err)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("couldn't read url: %v", err)
		return
	}

	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("couldn't parse html: %v", err)
		return
	}

	var allLinks []string
	for n := range doc.Descendants() {
		if n.Type == html.ElementNode && n.Data == "a" {
			links := findLinks(n.Attr)
			allLinks = append(allLinks, links...)
		}
	}

	fmt.Println(allLinks)
}

func findLinks(attrs []html.Attribute) []string {
	var links []string
	for _, attr := range attrs {
		if attr.Key == "href" {
			links = append(links, attr.Val)
		}
	}

	return links
}
