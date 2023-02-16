# Используем официальный образ Golang в качестве основы
FROM golang:latest

# Копируем исходный код в контейнер
WORKDIR /app
COPY . .

# Обновляем пакеты и устанавливаем зависимости для Google Sheets API
RUN apt update
RUN go mod download
# RUN go install google.golang.org/api/sheets/v
# RUN go install golang.org/x/oauth2/google
# RUN go install github.com/gocolly/colly

# Собираем приложение
RUN go build -o app

# Запускаем приложение
CMD ["./app"]