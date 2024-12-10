package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly/v2"
	"log"
	"os"
	"strings"
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

	// Открытие файла для записи результатов
	outputFile, err := os.Create("output.csv")
	if err != nil {
		log.Fatal("Unable to create output file: ", err)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	writer.Comma = '\t' //  разделитель
	defer writer.Flush()

	// Запись заголовков в CSV
	errWr := writer.Write([]string{"username", "message"})
	if err != nil {
		log.Fatalf("Error writing record to csv: %v", errWr)
		return
	}

	// Инициализация коллектора и настройка
	c := colly.NewCollector()

	// Добавление обработчика для каждого найденного сообщения
	c.OnHTML("div.post_wrap", func(e *colly.HTMLElement) {
		postWrap := e.DOM
		username := postWrap.Find("div.user_details span[itemprop='name']").Text()
		message := postWrap.Find("div[itemprop='commentText']").Text()

		fmt.Printf("author: %s\nmessage: %s\n", strings.TrimSpace(username), strings.TrimSpace(message))
		err := writer.Write([]string{strings.TrimSpace(username), strings.TrimSpace(message)})
		if err != nil {
			log.Fatalf("Error writing record to csv: %v", errWr)
			return
		}
	})

	c.OnHTML("div.topic_controls:not(.ipsPad_top_bottom_half) a[rel='next']", func(e *colly.HTMLElement) {
		nextPage := e.Attr("href")
		if nextPage != "" {
			fmt.Println("visiting next page: ", nextPage)
			err := e.Request.Visit(nextPage)
			if err != nil {
				log.Fatalf("Error visiting next page: %v", err)
				return
			}
		}
	})

	// Обработчик ошибок
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Обработка каждой ссылки
	for _, link := range links {
		err := c.Visit(link)
		fmt.Println("Visiting URL:", link)
		if err != nil {
			log.Printf("Error visiting %s: %v\n", link, err)
		}
	}
}
