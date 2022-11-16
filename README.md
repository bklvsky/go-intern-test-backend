# Микросервис для работы с балансом пользователей
## Реализовано:
1. Сценарий начисления средств на баланс
2. Сценарий резервирования средств
3. Сценарий одобрения операции
4. Сценарий получения баланса пользователя
5. Сценарий перевода средств между пользователями
6. Доп. задания:
    - Сценарий для получения истории последних операций над балансом пользователя
    - Сценарий разрезервирования денег

## Запуск
1. `git clone https://github.com/bklvsky/avito-intern-test-backend.git avito-user-balance`
2. `cd avito-user-balance`
3. `docker-compose pull`  
3. `docker-compose up --build`  

## Реализация
- База данных - postgres
- Фреймворк для Rest API - gorilla
- Данные запросов проходят валидацию

## Структура проекта
```
.
├──── Dockerfile  
├──── README.md  
├──── docker-compose.yml  
├──── go.mod  
├──── go.sum  
├──── cmd  
│     └──── main.go   -- точка входа в программу  
├──── db  
│     └──── postgres  
│           └──── initDB.go   -- подключение к БД  
├──── handlers -- функционал для работы с RestApi  
│     ├──── AppHandler.go  
│     ├──── ErrorHandler.go  
│     └──── UserHandler.go  
├──── models   -- структуры предметной области проекта  
│     ├──── Transaction.go  
│     └──── User.go  
│
├──── repositories   -- функционал для работы с БД  
│     └──── postgres  
│           ├──── TransactionsRepository.go  
│           └──── UsersRepository.go  
├──── resources  
│     └──── sql  
│           └──── init.sql   -- инициализация БД  
└──── validate   -- валидация json  
      └──── Validate.go  
```
