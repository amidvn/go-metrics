# Сервис сбора метрик и алертинга

Учебный проект по курсу "Go-разработчик".

Реализация сервера для сбора рантайм-метрик, который будет собирать репорты от агентов по протоколу HTTP.

## Начало работы

Для запуска на локальном компьютере необходимо:
* склонировать проект на свой компьютер <code>git clone https://github.com/amidvn/go-metrics.git</code>
* перейти в каталог проекта <code>cd go-metrics</code>

Запуск сервера: <code>go run .\cmd\server\main.go</code>

Запуск клиента: <code>go run .\cmd\client\main.go</code>

## Примеры

Пример запроса к серверу:

<code>POST /update/counter/someMetric/527 HTTP/1.1
Host: localhost:8080
Content-Length: 0
Content-Type: text/plain </code>

Пример ответа от сервера:

<code>HTTP/1.1 200 OK
Date: Tue, 21 Feb 2023 02:51:35 GMT
Content-Length: 11
Content-Type: text/plain; charset=utf-8</code>
