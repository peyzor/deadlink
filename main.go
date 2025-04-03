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
	"sync"
	"time"
)

type website struct {
	url string

	mu           sync.Mutex
	visitedLinks map[string]bool
	wg           sync.WaitGroup
	logs         chan string
}

func (w *website) visit(url string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.visitedLinks[url] = true
}

func (w *website) isVisited(url string) (visited bool, ok bool) {
	w.mu.Lock()
	w.mu.Unlock()

	visited, ok = w.visitedLinks[url]
	return
}

func (w *website) crawlPage(url string) error {
	body, err := w.requestPage(url)
	if err != nil {
		return err
	}

	node, err := parseHTML(body)
	if err != nil {
		return err
	}

	w.findLinks(node)
	return nil
}

func (w *website) findLinks(node *html.Node) {
	for n := range node.Descendants() {
		if n.Type != html.ElementNode || n.Data != "a" {
			continue
		}
		for _, attr := range n.Attr {
			if attr.Key != "href" {
				continue
			}
			link := attr.Val
			if _, ok := w.isVisited(link); !ok {
				w.visit(link)

				if strings.HasPrefix(link, "http") {
					continue
				}

				link = w.url + link

				w.wg.Add(1)
				go func() {
					defer w.wg.Done()
					err := w.crawlPage(link)
					if err != nil {
						log.Println(err)
					}
				}()
			}
		}
	}
}

func (w *website) requestPage(url string) ([]byte, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("couldn't get url: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	w.logs <- fmt.Sprintf("[%s] %s", resp.Status, url)

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

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("incorrect number of given arguments expected 1 but got: %d", len(os.Args)-1)
		return
	}

	site := website{
		url:          strings.TrimSuffix(os.Args[1], "/"),
		visitedLinks: make(map[string]bool),
		logs:         make(chan string),
	}
	site.visit(site.url)

	go func() {
		for l := range site.logs {
			fmt.Println(l)
		}
	}()

	site.wg.Add(1)
	go func() {
		defer site.wg.Done()

		err := site.crawlPage(site.url)
		if err != nil {
			log.Println(err)
		}
	}()

	site.wg.Wait()
	close(site.logs)

	links := slices.Collect(maps.Keys(site.visitedLinks))
	fmt.Println(links)
}
