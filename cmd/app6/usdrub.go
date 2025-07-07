package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"golang.org/x/text/encoding/charmap"
)

type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valutes []Valute `xml:"Valute"`
}

type Valute struct {
	CharCode string `xml:"CharCode"`
	Nominal  int    `xml:"Nominal"`
	Name     string `xml:"Name"`
	Value    string `xml:"Value"`
}

func main() {
	// Настраиваем логирование
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Ошибка открытия файла логов:", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			//
		}
	}(file)
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.Println("Программа запущена")

	log.Println("Запуск системного трея")
	systray.Run(onReady, onExit)
}

func onReady() {
	log.Println("Функция onReady вызвана")
	systray.SetTitle("USD: Загрузка...")
	systray.SetTooltip("Курс доллара США (ЦБ РФ)")

	mRefresh := systray.AddMenuItem("Обновить", "Обновить курс валют")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Выход", "Закрыть приложение")
	log.Println("Меню инициализировано")

	go func() {
		log.Println("Запуск горутины обновления курса")
		updateExchangeRate()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-mRefresh.ClickedCh:
				log.Println("Ручное обновление курса")
				updateExchangeRate()
			case <-ticker.C:
				log.Println("Автоматическое обновление курса")
				updateExchangeRate()
			}
		}
	}()

	go func() {
		<-mQuit.ClickedCh
		log.Println("Команда выхода получена")
		systray.Quit()
	}()
}

func onExit() {
	log.Println("Приложение завершено")
}

func updateExchangeRate() {
	log.Println("Обновление курса...")
	url := "https://www.cbr.ru/scripts/XML_daily.asp"
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			log.Println("Перенаправление на:", req.URL.String())
			return nil
		},
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Ошибка создания запроса:", err)
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка создания запроса: " + err.Error())
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Ошибка HTTP-запроса:", err)
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка получения данных: " + err.Error())
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("Ошибка закрытия HTTP-соединения:", err)
		}
	}()

	// Проверяем Content-Type
	contentType := resp.Header.Get("Content-Type")
	log.Printf("Content-Type ответа: %s", contentType)
	if !strings.Contains(contentType, "xml") {
		log.Println("Ответ не является XML")
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка: ответ не является XML")
		return
	}

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка чтения данных:", err)
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка чтения данных: " + err.Error())
		return
	}

	// Логируем первые 500 символов сырого ответа
	// log.Printf("Сырые данные (первые 500 символов): %s", string(body[:minimum(len(body), 500)]))

	// Декодируем из Windows-1251 в UTF-8
	decoder := charmap.Windows1251.NewDecoder()
	bodyUTF8, err := decoder.Bytes(body)
	if err != nil {
		log.Println("Ошибка декодирования Windows-1251:", err)
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка декодирования данных: " + err.Error())
		return
	}

	// Удаляем BOM, если он есть
	bodyUTF8 = bytes.TrimPrefix(bodyUTF8, []byte{0xEF, 0xBB, 0xBF})

	// Заменяем encoding="windows-1251" на encoding="UTF-8"
	bodyUTF8 = bytes.Replace(bodyUTF8, []byte(`encoding="windows-1251"`), []byte(`encoding="UTF-8"`), 1)

	// Логируем преобразованные данные
	// log.Printf("Преобразованные данные (первые 500 символов): %s", string(bodyUTF8[:min(len(bodyUTF8), 500)]))

	var valCurs ValCurs
	if err := xml.Unmarshal(bodyUTF8, &valCurs); err != nil {
		log.Println("Ошибка парсинга XML:", err)
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка парсинга XML: " + err.Error())
		return
	}

	for _, valute := range valCurs.Valutes {
		if valute.CharCode == "USD" {
			log.Printf("Курс USD найден: %s на %s", valute.Value, valCurs.Date)
			systray.SetTitle(fmt.Sprintf("USD: %s ₽", valute.Value))
			systray.SetTooltip(fmt.Sprintf("Курс доллара США (ЦБ РФ) на %s: %s ₽", valCurs.Date, valute.Value))
			return
		}
	}

	log.Println("Курс USD не найден")
	systray.SetTitle("USD: Не найдено")
	systray.SetTooltip("Курс USD не найден в данных ЦБ РФ")
}

func minimum(a, b int) int {
	if a < b {
		return a
	}
	return b
}
