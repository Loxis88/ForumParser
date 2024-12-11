package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	// Открытие файла с ссылками
	file, err := os.Open("temaLinks.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Чтение URL-ов из файла
	var links []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		links = append(links, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Инициализация коллектора и настройка
	c := colly.NewCollector(colly.Async(true))

	errLim := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 20,
		Delay:       100 * time.Millisecond,
	})
	if errLim != nil {
		log.Fatal(errLim)
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results [][]string // Срез для хранения результатов

	// Добавление обработчика для каждого найденного сообщения
	c.OnHTML("div.post_wrap", func(e *colly.HTMLElement) {
		defer wg.Done() // Уменьшаем счетчик после завершения обработки

		postWrap := e.DOM

		username := postWrap.Find("div.user_details span[itemprop='name']").Text()
		message := postWrap.Find("div[itemprop='commentText']:not(blockquote.ipsBlockquote)").Text()

		mu.Lock()
		results = append(results, []string{strings.TrimSpace(username), strings.TrimSpace(message)})
		mu.Unlock()
	})

	c.OnHTML("div.topic_controls:not(.ipsPad_top_bottom_half) a[rel='next']", func(e *colly.HTMLElement) {
		nextPage := e.Attr("href")
		if nextPage != "" {
			wg.Add(1) // Увеличиваем счетчик перед вызовом Visit
			err := e.Request.Visit(nextPage)
			if err != nil {
				log.Fatalf("Error visiting next page: %v", err)
			}
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	for _, link := range links {
		wg.Add(1) // Увеличиваем счетчик перед вызовом Visit
		err := c.Visit(link)
		if err != nil {
			log.Printf("Error visiting %s: %v\n", link, err)
		}
	}

	c.Wait() // Ожидание завершения всех асинхронных запросов

	// Открытие файла для записи результатов
	outputFile, err := os.Create("output.csv")
	if err != nil {
		log.Fatal("Unable to create output file: ", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	writer.Comma = '\t' // разделитель
	defer writer.Flush()

	// Запись заголовков в CSV
	errWr := writer.Write([]string{"username", "message"})
	if errWr != nil {
		log.Fatalf("Error writing record to csv: %v", errWr)
		return
	}

	// Запись результатов в CSV
	for _, result := range results {
		errWr := writer.Write(result)
		if errWr != nil {
			log.Fatalf("Error writing record to csv: %v ", errWr)
		}
	}

	fmt.Println("\nResults have been written to output.csv")
}
