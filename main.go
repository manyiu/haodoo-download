package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL        = "http://www.haodoo.net"
	bookPathPrefix = "?M=book&P="
	downPathPrefix = baseURL + "/" + "?M=d&P="
)

type book struct {
	author string
	title string
	link string
	format string
}

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
			}
		})

		close(l)
	}()

	return l
}

func getDownloadURL(p <-chan string, f string) <-chan book {
	l := make(chan book)

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

			doc.Find("input[onclick^=Download" + strings.Title(f) + "]").Each(func(i int, s *goquery.Selection){
				o, _ := s.Attr("onclick")

				p := s.Parent()
				ft := p.Find("font").First()
				a := ft.Text()
				t, _ := goquery.OuterHtml(p)

				rd := regexp.MustCompile("(?:Download" + strings.Title(f) + "\\(')([0-9a-zA-Z]{1,})(?:'\\))")
				rt := regexp.MustCompile("(?:</font>《)(\\S{1,})(?:》<input )")


				sd := rd.FindStringSubmatch(o)
				st := rt.FindStringSubmatch(t)
				
				if len(sd) == 2 && len(st) == 2 {
					fmt.Println(sd[1], st[1], a)
					l <- book{
						author: a,
						title: st[1],
						link: sd[1],
						format: f,
					}
				}
			})
		}

		close(l)
	}()

	return l
}

func main() {
	var u, f = os.Args[1], os.Args[2]

	p := getBookDownloadPage(u)

	d := getDownloadURL(p, strings.ToLower(f))

	for v := range d {
		fmt.Println(v)
	}
}
