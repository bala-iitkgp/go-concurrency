package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type ConcurrentSet struct {
	mu sync.Mutex
	s  map[string]bool
}

func (set *ConcurrentSet) GetValue(url string) bool {
	set.mu.Lock()
	defer set.mu.Unlock()
	_, ok := set.s[url]
	return ok
}

func (set *ConcurrentSet) InsertValue(url string) {
	set.mu.Lock()
	set.s[url] = true
	set.mu.Unlock()
	return
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, set *ConcurrentSet) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		return
	}

	if !set.GetValue(url) {
		set.InsertValue(url)
		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			fmt.Println("err: ", err)
			return
		}

		fmt.Printf("found: %s %q\n", url, body)
		var wg sync.WaitGroup
		for _, u := range urls {
			//Add should happen in the main thread
			wg.Add(1)
			go func(u string) {
				// Done should happen in the child thread
				// using defer ensures wg.Done() is called
				// even if there is a crash in the below code
				defer wg.Done()
				Crawl(u, depth-1, fetcher, set)
			}(u)
		}
		wg.Wait()
	}
	return
}

func main() {
	s := ConcurrentSet{s: make(map[string]bool)}
	Crawl("https://golang.org/", 4, fetcher, &s)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
