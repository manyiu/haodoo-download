package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const (
	baseURL        = "http://www.haodoo.net"
	bookPathPrefix = "?M=book&P="
	bookDownloadPathPrefix = "?M=d&P="
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

func downloadBook(b <- chan book) {
	wd, err := os.Getwd()
	var wg sync.WaitGroup

	if err != nil {
		log.Fatal(err)
	}

	for v := range b {
		wg.Add(1)
		go func(v book) {
			defer wg.Done()

			res, err := http.Get(baseURL + "/" + bookDownloadPathPrefix + v.link + "." + v.format)
			
			if err != nil {
				return
			}
			
			defer res.Body.Close()
			
			p := filepath.Join(wd, v.author)
			fn := v.title + "." + v.format
			
			err = os.MkdirAll(p, os.ModePerm)

			if err != nil {
				return
			}

			out, err := os.Create(filepath.Join(p, fn))
			
			if err != nil {
				return
			}
			
			defer out.Close()
			
			_, err = io.Copy(out, res.Body)
		}(v)
	}

	wg.Wait()
}

func main() {
	var u, f = os.Args[1], os.Args[2]

	p := getBookDownloadPage(u)

	d := getDownloadURL(p, strings.ToLower(f))

	downloadBook(d)
}
