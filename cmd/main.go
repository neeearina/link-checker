package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Link struct {
	URL             string
	IsValidURL      bool
	IsAccessibleURL bool
	Error           string
}

func readLinks() []Link {
	file, err := os.Open("example.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var links []Link

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "http//") {
			line = strings.Replace(line, "http//", "http://", 1)
		}
		link := Link{URL: line}
		links = append(links, link)
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return links
}

func checkSingleLink(link Link) Link {
	result := link

	_, err := url.Parse(link.URL)
	if err != nil {
		result.IsValidURL = false
		result.IsAccessibleURL = false
		result.Error = fmt.Sprintf("Некорректный URL: %v", err)
		return result
	}

	result.IsValidURL = true

	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get(link.URL)
	if err != nil {
		result.IsAccessibleURL = false
		result.Error = fmt.Sprintf("Ошибка доступа: %v", err)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.IsAccessibleURL = true
	} else {
		result.IsAccessibleURL = false
		result.Error = fmt.Sprintf("Сервер вернул статус: %d", resp.StatusCode)
	}

	return result
}

func checkLinksParallel(links []Link) []Link {
	results := make(chan Link, len(links))
	var wg sync.WaitGroup

	for _, link := range links {
		wg.Add(1)
		go func(l Link) {
			defer wg.Done()
			results <- checkSingleLink(l)
		}(link)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var checkedLinks []Link
	for result := range results {
		checkedLinks = append(checkedLinks, result)
	}

	return checkedLinks
}

func main() {
	links := readLinks()

	start := time.Now()

	checkedLinks := checkLinksParallel(links)

	elapsed := time.Since(start)
	fmt.Printf("\nПроверка выполнена за %v\n", elapsed)

	fmt.Println("\nРезультаты проверки ссылок:")
	fmt.Println("------------------------")
	for _, link := range checkedLinks {
		if !link.IsValidURL {
			fmt.Printf("❓ %s\n   %s\n", link.URL, link.Error)
			continue
		}

		if link.IsAccessibleURL {
			fmt.Printf("✅ %s\n", link.URL)
		} else {
			fmt.Printf("❌ %s\n   %s\n", link.URL, link.Error)
		}
	}
	fmt.Println()
}