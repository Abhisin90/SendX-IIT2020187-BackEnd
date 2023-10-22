package main

import (
	"math/rand"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"github.com/gocolly/colly"
	"time"
)

// The function "home" parses and executes an HTML template file named "index.html" and writes the
// output to the http.ResponseWriter.
func home(w http.ResponseWriter, r *http.Request){
	var fileName = "index.html"
	t,err := template.ParseFiles(fileName)
	if err != nil {
		fmt.Println("Error occured in parsing file", err)
		return
	}
	err = t.ExecuteTemplate(w,fileName,nil)
	if err != nil {
		fmt.Println("Error occured during execution of file",err)
		return
	}
}

var UserAgents = []string {
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41",
    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1",
}
// The function `crawling` uses the Colly library in Go to crawl web pages starting from a given URL.
func crawling(url string) error {	
	// defer wg.Done()
	c := colly.NewCollector()
	c.UserAgent = UserAgents[rand.Int() % len(UserAgents)]
	fmt.Println(c.UserAgent)
	c.OnRequest(func(r *colly.Request) { 
		fmt.Println("Visiting", r.URL) 	
	}) 
	c.OnError(func(_ *colly.Response, err error) { 
		log.Println("Something went wrong:", err) 
	}) 
	c.OnResponse(func(r *colly.Response) { 
		fmt.Println("Visited", r.Request.URL) 
	}) 	
    c.OnHTML("a[href]", func(e *colly.HTMLElement) { 
		e.Request.Visit(e.Attr("href"))
	})

	fmt.Println("Starting crawl at: ", url) 
		
	if err := c.Visit(url); err != nil { 
	    fmt.Println("Error on start of crawl: ", err) 
		return err
	} 
	c.Wait()
	return nil 
}

// The function `crawlurl` takes in a URL and a flag indicating if the customer is paid or not, and
// then adds the URL to a list of pages to be crawled, and starts a certain number of goroutines for
// crawling based on the customer's paid status. 
func crawlurl(w http.ResponseWriter, r *http.Request){
	url := r.FormValue("url")
	crawling(url)
	// flag := r.FormValue("paid")
    // var wg sync.WaitGroup
	// isPaidCustomer := false
	// if flag == "true" {
	// 	isPaidCustomer = true
	// }
	retries := 3

    // Set the delay between retries.
    delay := 100 * time.Millisecond

    // Retry the function until it succeeds.
    for i := 0; i < retries; i++ {
        err := crawling(url)
        if err != nil {
            fmt.Println("Error:", err)
            time.Sleep(delay)
            continue
        }

        // The function succeeded.
        break
    }

	// if(isPaidCustomer){
	// 	for i := 0; i < 5; i++ {
	// 		wg.Add(1)
	// 		go crawling(&wg,url)
	// 	}

	// } else {
	// 	for i :=0; i < 2; i++ {
	// 		wg.Add(1)
	// 		go crawling(&wg,url)
	// 	}
	// }
	// wg.Wait()
}

// The handler function routes incoming HTTP requests to different 
// functions based on the requested URL path.
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/home":
		home(w,r)
	case "/crawl-url":
		crawlurl(w,r)
	}
}

func main(){
	http.HandleFunc("/",handler)
	http.ListenAndServe(":8080",nil)
}

