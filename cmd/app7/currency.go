package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
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

type CryptoCurrency struct {
	Bitcoin  map[string]float64 `json:"bitcoin"`
	Ethereum map[string]float64 `json:"ethereum"`
}

var (
	currentCurrency = "USD" // Текущая валюта по умолчанию
	mu              sync.Mutex
)

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
	systray.SetTooltip("Курс валют")

	// Создаем подменю для выбора валюты
	mCurrency := systray.AddMenuItem("Валюта", "Выбрать валюту")
	mUSD := mCurrency.AddSubMenuItem("USD", "Доллар США")
	mBTC := mCurrency.AddSubMenuItem("BTC", "Bitcoin")
	mETH := mCurrency.AddSubMenuItem("ETH", "Ethereum")
	systray.AddSeparator()
	mRefresh := systray.AddMenuItem("Обновить", "Обновить курс валют")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Выход", "Закрыть приложение")
	log.Println("Меню инициализировано")

	// Выполняем начальное обновление синхронно
	updateExchangeRate()

	go func() {
		log.Println("Запуск горутины обновления курса")
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-mRefresh.ClickedCh:
				log.Println("Ручное обновление курса")
				mu.Lock()
				systray.SetTitle(fmt.Sprintf("%s: Загрузка...", currentCurrency))
				mu.Unlock()
				updateExchangeRate()
			case <-mUSD.ClickedCh:
				log.Println("Переключение на USD")
				mu.Lock()
				currentCurrency = "USD"
				systray.SetTitle("USD: Загрузка...")
				mu.Unlock()
				updateExchangeRate()
			case <-mBTC.ClickedCh:
				log.Println("Переключение на BTC")
				mu.Lock()
				currentCurrency = "BTC"
				systray.SetTitle("BTC: Загрузка...")
				mu.Unlock()
				updateExchangeRate()
			case <-mETH.ClickedCh:
				log.Println("Переключение на ETH")
				mu.Lock()
				currentCurrency = "ETH"
				systray.SetTitle("ETH: Загрузка...")
				mu.Unlock()
				updateExchangeRate()
			case <-ticker.C:
				log.Println("Автоматическое обновление курса")
				mu.Lock()
				systray.SetTitle(fmt.Sprintf("%s: Загрузка...", currentCurrency))
				mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
	log.Printf("Обновление курса для %s...", currentCurrency)
	switch currentCurrency {
	case "USD":
		updateUSDRate()
	case "BTC", "ETH":
		updateCryptoRate()
	default:
		log.Println("Неизвестная валюта")
		systray.SetTitle("Ошибка")
		systray.SetTooltip("Ошибка: неизвестная валюта")
	}
}

func updateUSDRate() {
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			//
		}
	}(resp.Body)

	if !strings.Contains(resp.Header.Get("Content-Type"), "xml") {
		log.Println("Ошибка: ответ не является XML")
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка: ответ не является XML")
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Ошибка чтения данных:", err)
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка чтения данных: " + err.Error())
		return
	}

	decoder := charmap.Windows1251.NewDecoder()
	bodyUTF8, err := decoder.Bytes(body)
	if err != nil {
		log.Println("Ошибка декодирования Windows-1251:", err)
		systray.SetTitle("USD: Ошибка")
		systray.SetTooltip("Ошибка декодирования данных: " + err.Error())
		return
	}

	bodyUTF8 = bytes.TrimPrefix(bodyUTF8, []byte{0xEF, 0xBB, 0xBF})
	bodyUTF8 = bytes.Replace(bodyUTF8, []byte(`encoding="windows-1251"`), []byte(`encoding="UTF-8"`), 1)

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

func updateCryptoRate() {
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin,ethereum&vs_currencies=rub"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Println("Ошибка HTTP-запроса для криптовалют:", err)
		systray.SetTitle(fmt.Sprintf("%s: Ошибка", currentCurrency))
		systray.SetTooltip(fmt.Sprintf("Ошибка получения данных %s: %s", currentCurrency, err.Error()))
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			//
		}
	}(resp.Body)

	var price CryptoCurrency
	if err := json.NewDecoder(resp.Body).Decode(&price); err != nil {
		log.Println("Ошибка парсинга JSON для криптовалют:", err)
		systray.SetTitle(fmt.Sprintf("%s: Ошибка", currentCurrency))
		systray.SetTooltip(fmt.Sprintf("Ошибка парсинга данных %s: %s", currentCurrency, err.Error()))
		return
	}

	var value float64
	var currencyName string
	switch currentCurrency {
	case "BTC":
		value = price.Bitcoin["rub"]
		currencyName = "Bitcoin"
	case "ETH":
		value = price.Ethereum["rub"]
		currencyName = "Ethereum"
	}

	if value == 0 {
		log.Printf("Курс %s не найден", currentCurrency)
		systray.SetTitle(fmt.Sprintf("%s: Не найдено", currentCurrency))
		systray.SetTooltip(fmt.Sprintf("Курс %s не найден", currencyName))
		return
	}

	log.Printf("Курс %s найден: %.2f RUB", currentCurrency, value)
	systray.SetTitle(fmt.Sprintf("%s: %.2f ₽", currentCurrency, value))
	systray.SetTooltip(fmt.Sprintf("Курс %s на %s: %.2f ₽", currencyName, time.Now().Format("01.01.2002"), value))
}
