package main

import (
	"sync"
	"strconv"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"log"
)

type Proxy struct {
	ip string
	port string
	proxyType int //1-高匿 2-匿名 3-普通
	connectType int //1-http 2-https 3-socks

}

var chanProxy = make(chan Proxy, 10)

func Crawl(){

}

func crawlXici(){
	var waitSync sync.WaitGroup
	client := &http.Client{}

	for i := 1; i < 2; i++ {
		waitSync.Add(1)
		go func(i int) {
			defer waitSync.Done()
			urlXici :=  "https://www.xicidaili.com/nn/" + strconv.Itoa(i)
			req, _ := http.NewRequest("GET", urlXici, nil)
			req.Header.Set("Host", "www.xicidaili.com")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
				return
			}
			if resp.StatusCode == http.StatusOK {
				doc, _ := goquery.NewDocumentFromReader(resp.Body)
				table := doc.Find("#ip_list")
				table.Find("tr").Each(func(i int, selection *goquery.Selection) {
					if i == 0 {
						return
					}
					row := selection.Find("td")
					ip := row.Eq(1).Text()
					port := row.Eq(2).Text()
					ct := row.Eq(2).Text()
					var connectType int
					if ct == "HTTP" {
						connectType = 1
					}else if ct == "HTTPS" {
						connectType = 2
					}else{
						connectType = 3
					}
					chanProxy <- Proxy{ip:ip,port:port,proxyType:1,connectType:connectType}
				})

			}

		}(i)
	}
	waitSync.Wait()
}

func main(){
	go crawlXici()
	var proxyList []Proxy
	for proxy := range chanProxy{
		proxyList = append(proxyList, proxy)
		println(proxy.ip)
	}


}
