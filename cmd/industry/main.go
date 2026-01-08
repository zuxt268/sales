package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	txt, err := getCompInfo("https://21lab.biz/")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(txt)
}

func getCompInfo(siteUrl string) (string, error) {
	u, err := url.Parse(siteUrl)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "service")
	fmt.Println(u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 404 {
		resp.Body.Close()
		resp, err = http.Get(u.String())
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	body := doc.Find("body")
	body.Find("script,style,link,noscript").Remove()

	text := strings.Join(strings.Fields(body.Text()), " ")
	return text, nil
}
