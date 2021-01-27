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
	bookPathPrefix = "?M=book&P="
	downPathPrefix = baseURL + "/?M=d&P="
)

func main() {
	var requestURL, format = os.Args[1], os.Args[2]

	fmt.Println(format)

	res, err := http.Get(requestURL)

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
				fmt.Println(val)
			}
		})
	}()
}
