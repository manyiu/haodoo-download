package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL        = "http://www.haodoo.net"
	bookPathPrefix = "/?M=book&P="
	downPathPrefix = baseURL + "/?M=d&P="
)

func getBookDownloadPage(url string) <-chan string {
	l := make(chan string)

	res, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if val, _ := s.Attr("href"); strings.Contains(val, bookPathPrefix) {
				l <- val
				fmt.Println(val)
			}
		})

		close(l)
	}()

	return l
}

func getDownloadUrl(p <-chan string) <-chan string {
	l := make(chan string)

	go func() {
		for v := range p {
			res, err := http.Get(baseURL + "/" + v)

			if err != nil {
				log.Fatal(err)
			}

			defer res.Body.Close()

			doc, err := goquery.NewDocumentFromResponse(res)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(doc)
		}

		close(l)
	}()

	return l
}

func main() {
	var requestURL, format = os.Args[1], os.Args[2]

	fmt.Println(format)

	p := getBookDownloadPage(requestURL)

	getDownloadUrl(p)
}
