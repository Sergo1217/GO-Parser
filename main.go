package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gocolly/colly"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

//Структура конфигв
type Data struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	Range         string `json:"range"`
}

//получает данные из файла config.json
func loadConfig() (*Data, error) {
	content, err := ioutil.ReadFile("./config.json")
	if err != nil {
		return nil, err
	}

	var payload Data
	err = json.Unmarshal(content, &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

// Получает токен, сохраняет его, затем возвращает сгенерированный клиент.
func getClient(config *oauth2.Config) *http.Client {
	// Файл token.json хранит маркеры доступа и обновления пользователя
	// и создается автоматически, когда поток авторизации завершается в первый раз.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Запрашивает токен из Интернета, затем возвращает полученный токен.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Перейдите по следующей ссылке в браузере и введите код авторизации: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Невозможно считать код авторизации: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Невозможно получить токен из Интернета: %v", err)
	}
	return tok
}

// Извлекает токен из локального файла.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Сохраняет token в файл.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Сохранение файла учетных данных в: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

//Парсит таблицу с сайта
func parseTable() ([][]interface{}, error) {
	// Создаем новый коллектор
	c := colly.NewCollector(colly.AllowedDomains("confluence.hflabs.ru"))

	// Парсим таблицу
	rows := [][]interface{}{}
	c.OnHTML("table", func(e *colly.HTMLElement) {
		// Находим заголовок таблицы
		header := []interface{}{}
		e.ForEach("thead tr th", func(_ int, el *colly.HTMLElement) {
			// Добавляем текст ячейки заголовка в слайс
			header = append(header, el.Text)
		})
		rows = append(rows, header)

		// Находим все строки таблицы
		e.ForEach("tbody tr", func(_ int, el *colly.HTMLElement) {
			row := []interface{}{}
			// Находим все ячейки в текущей строке
			el.ForEach("td", func(_ int, el2 *colly.HTMLElement) {
				// Добавляем текст ячейки в строку
				row = append(row, el2.Text)
			})
			rows = append(rows, row)
		})
	})

	// Загружаем страницу
	err := c.Visit("https://confluence.hflabs.ru/pages/viewpage.action?pageId=1181220999")
	if err != nil {
		return nil, err
	}
	return rows, nil
}

//Сохраняет данные в Google Таблицы
func updateGoogleSheet(spreadsheetId string, sheetRange string, rows [][]interface{}) error {
	// Инициализируем Google Sheets API
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
		return err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
		return err
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		return err
	}

	// Создаем новый запрос на обновление листа в таблице
	valueRange := &sheets.ValueRange{
		Values: rows,
	}
	_, err = srv.Spreadsheets.Values.Update(spreadsheetId, sheetRange, valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Fatalf("Unable to update data to sheet: %v", err)
		return err
	}

	fmt.Println("Data updated successfully")
	return nil
}

func main() {
	payload, err := loadConfig()
	if err != nil {
		log.Fatal("Ошибка при чтении конфига: ", err)
	}

	rows, err := parseTable()
	if err != nil {
		log.Fatal("Ошибка при парсинге таблицы: ", err)
	}

	err = updateGoogleSheet(payload.SpreadsheetID, payload.Range, rows)
	if err != nil {
		log.Fatal("Ошибка при загрузке таблицы: ", err)
	}
}
