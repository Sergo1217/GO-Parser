# Имя проекта
NAME=parser-go

# Команда для сборки приложения
build:
	go build -o $(NAME) .

# Команда для запуска приложения
run:
	go run .

# Команда для сборки Docker-образа
docker-build:
	docker build -t $(NAME) .

# Команда для запуска контейнера из Docker-образа
docker-run:
	docker run --rm $(NAME)

# Команда для удаления Docker-образа
docker-rm:
	docker rmi $(NAME)