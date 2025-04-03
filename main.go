package main

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"
)

var visitedLinks = make(map[string]bool)
var URL string

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("incorrect number of given arguments expected 1 but got: %d", len(os.Args)-1)
		return
	}

	URL = os.Args[1]
	visitedLinks[URL] = true

	URL = strings.TrimSuffix(URL, "/")
	err := crawlPage(URL)
	if err != nil {
		log.Println(err)
	}

	links := slices.Collect(maps.Keys(visitedLinks))
	fmt.Println(links)
}

func requestPage(url string) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("couldn't get url: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("couldn't read url: %v", err)
		return nil, err
	}

	return body, nil
}

func parseHTML(body []byte) (*html.Node, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		fmt.Printf("couldn't parse html: %v", err)
		return nil, err
	}

	return doc, nil
}

func findLinks(node *html.Node) {
	for n := range node.Descendants() {
		if n.Type != html.ElementNode || n.Data != "a" {
			continue
		}
		for _, attr := range n.Attr {
			if attr.Key != "href" {
				continue
			}
			link := attr.Val
			if _, ok := visitedLinks[link]; !ok {
				visitedLinks[link] = true

				if strings.HasPrefix(link, "http") {
					continue
				}

				link = URL + link
				err := crawlPage(link)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func crawlPage(url string) error {
	body, err := requestPage(url)
	if err != nil {
		return err
	}

	node, err := parseHTML(body)
	if err != nil {
		return err
	}

	findLinks(node)
	return nil
}
