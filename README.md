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
│
├──── Dockerfile
├──── README.md
├──── docker-compose.yml
├──── go.mod
├──── go.sum
│
├──── resources
│     └──── sql   -- инициализация БД
│           └──── init.sql
├──── cmd
│     └──── main.go   -- точка входа в программу
│
├──── db
│     └──── postgres
│         └──── initDB.go   -- подключение к БД
│
├──── handlers   -- функционал для работы с RestApi
│     ├──── appHandler.go
│     ├──── errorHandler.go
│     ├──── middleWare.go
│     └──── userHandler.go
│
├──── models    -- структуры предметной области проекта
│     └──── models.go
│   
├──── repositories   -- функционал для работы с БД
│     └──── postgres
│           ├──── transactionsRepository.go
│           └──── usersRepository.go
│
└──── validate   -- валидация json
      └──── validate.go
```

## Примеры запросов
### Вернуть баланс пользователя
- Запрос Postman:

```
[GET] localhost:8080/users/1  
```
- Тело ответа:  
```
{
    "id": 3,
    "balance": 100.5
}
```

### Создание нового пользователя или пополнение баланса у существующего
- Запрос Postman:
```
[POST] localhost:8080/users/
```
```
{
    "id": 3,
    "balance": 100.5
}
```
- Тело ответа:  
```
{
    "status": "successful"
}
```
### Резервирование средств

- Запрос Postman:
```
[POST] localhost:8080/orders/
```
```
{
    "orderId": 2,
    "clientId": 3,
    "serviceId": 2,
    "value": 100,
    "status": "in process" // на этапе резервирования денег - опциональное поле
}
```
- Тело ответа:
```
{
    "status": "successful"
}
```
### Подтверждение списания денег:
- Запрос Postman:
```
[POST] localhost:8080/orders/
```
```
{
    "orderId": 2,
    "clientId": 3,
    "serviceId": 2,
    "value": 100,
    "status": "approved" // обязательное поле
}
```
- Тело ответа:  
```
{
    "status": "successful"
}
```
### Разрезервирование денег
- Запрос Postman:
```
[POST] localhost:8080/orders/
```
```
{
    "orderId": 2,
    "clientId": 3,
    "serviceId": 2,
    "value": 100,
    "status": "canceled" // обязательное поле
}
```
- Тело ответа:  
```
{
    "status": "successful"
}
```
### Перевод средств между пользователями
- Запрос Postman:
```
[POST] localhost:8080/transfer
```
```
{
    "sender": 2,
    "recipient": 3,
    "value": 20
}
```
- Тело ответа:  
```
{
    "status": "successful"
}
```
### История баланса пользователя:
- В качестве параметров принимает номер ID пользователя, номер страницы и тип сортировки: "by_date" или "by_value" 
- Количество записей на странице - 10
- Запрос Postman:
```
[GET] localhost:8080/history/
```
```
{
    "userId": 1,
    "page": 1,
    "sort": "by_date"
}
```
- Тело ответа:  
```
{
    "history": [
        {
            "orderId": 3,
            "clientId": 1,
            "serviceId": 2,
            "value": 10,
            "time": "2022-11-16T17:48:44.979075Z",
            "note": "Order 3 canceled"
        },
        {
            "orderId": 3,
            "clientId": 1,
            "serviceId": 2,
            "value": -10,
            "time": "2022-11-16T17:48:33.888109Z"
        }
    ]
}
```


### Ответ в случае невалидных данных:
- Выставляется соответствующий статус ответа
- Ответ в формате JSON по такому шаблону:
```
{
    "status": "failed", // обязательное поле
    "message": "invalid character '}' looking for beginning of object key string", // обязательное поле с описанием ошибки
    "description": "Error while decoding user JSON" // опциональное поле, уточняющее обстоятельства ошибки
}
```
- Примеры:
1. Пользователь не найден:
```
{
    "status": "failed",
    "message": "No User with ID 5 found"
}
```
2. Отмена несуществующего заказа:
```
{
    "status": "failed",
    "message": "Order doesn't exist. It can't be created with canceled status."
}
```
3. Изменение завершенного заказа:
```
{
    "status": "failed",
    "message": "Order is already approved and can't be modified (canceled)"
}
```
4. Неудачная попытка списания средств:
```
{
    "status": "failed",
    "message": "Not enough money in the account"
}
```
### О дополнительных проектных решениях:
- В отличие от данных пользователей, я решила не проводить валидацию ID заказов и услуг и не взаимодействовать с их базами данных. В ТЗ задания указано, что микросервис должен предназначаться исключительно для манипуляции с балансом пользователей, и я исхожу из того, что эти данные будут поступать валидными, так как они управляются другим сервисом.
- При обращении к целой категории, принимаются адреса как в формате
```/users/```, так и ```/users```
- Предоставляются дополнительные HTTP методы для внутреннего пользования:
    - ```GET /users/```  
    Предоставляет всех пользователей с данными об их балансе  
    - ```GET /orders/```  
    Предоставляет список всех транзакций  
    - ```GET /orders/{id}```  
    Предоставляет данные о последней транзакции по ID заказа
