package main

import (
	"sync"
	"strconv"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"log"
	"github.com/gomodule/redigo/redis"
	"strings"
	"net/url"
	"crypto/tls"
)
var rc redis.Conn


type Proxy struct {
	ip string
	port string
	proxyType int //1-高匿 2-匿名 3-普通
	connectType int //1-http 2-https 3-socks
}

func (p *Proxy) proxyToStr() string{
	return p.ip + "|" + p.port + "|" + strconv.Itoa(p.proxyType) + "|" + strconv.Itoa(p.connectType)
}

func strToProxy(s string) Proxy {
	l := strings.Split(s, "|")
	proxyType, _  := strconv.Atoi(l[2])
	connectType, _  := strconv.Atoi(l[3])
	return Proxy{l[0], l[1], proxyType, connectType}
}

var chanProxy = make(chan Proxy)

func Crawl(){

}

func crawlXici(){
	var waitSync sync.WaitGroup
	var client *http.Client
	px, err := getRandomProxy()
	if err == nil {
		sc := ""
		if px.connectType == 1 {
			sc = "http://"
		}else if px.connectType == 2 {
			sc = "https://"
		}
		proxyUrl, _ := url.Parse(sc + px.ip + ":" + px.port)
		tr := &http.Transport{
			Proxy:	http.ProxyURL(proxyUrl),
			TLSClientConfig:	&tls.Config{InsecureSkipVerify:true},
		}
		client = &http.Client{
			Transport:	tr,
		}
		log.Println("[proxyUrl]" , sc + px.ip + ":" + px.port)
	}else{
		client = &http.Client{}
	}
	for i := 1; i < 5; i++ {
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
	close(chanProxy)
}

func saveProxy() {
	for proxy := range chanProxy {
		str := proxy.proxyToStr()
		log.Println("[save]", str)
		rc.Do("SADD", "proxyList", str)
	}
}

func getRandomProxy() (Proxy Proxy, err error){
	v, err := redis.String(rc.Do("SRANDMEMBER", "proxyList"))
	if err != nil {
		return Proxy,err
	}else {
		return strToProxy(v), err
	}

}

func main(){
	var err error
	rc, err = redis.Dial("tcp", "127.0.0.1:6379")

	defer rc.Close()
	if err != nil {
		log.Panic("redis connection error")
	}
	go crawlXici()
	saveProxy()


}
