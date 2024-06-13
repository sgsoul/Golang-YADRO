# РАЗРАБОТКА МИКРОСЕРВИСНЫХ ПРИЛОЖЕНИЙ НА GOLANG

Про курс посмотреть [тут](https://careers.yadro.com/practical-courses/golang/)

-   [Нормализация поисковых запросов. Стемминг](https://github.com/sgsoul/golangYADRO/tree/Normalization-of-search-queries-Stemming)
-   [Работа с REST API](https://github.com/sgsoul/golangYADRO/tree/REST-API)
-   [Оптимизация производительности приложения на Go: многопоточность, конкурентный доступ, кэширование](https://github.com/sgsoul/golangYADRO/tree/Performance-optimization)
-   [Индексация, поиск и ранжирование](https://github.com/sgsoul/golangYADRO/tree/Indexing-search-ranking)
-   [Создание и тестирование веб-сервиса](https://github.com/sgsoul/golangYADRO/tree/Web-service)
-   [Основы работы с SQL. Схемы, миграции](https://github.com/sgsoul/golangYADRO/tree/SQL-database)
-   [Ограничение доступа к сервису](https://github.com/sgsoul/golangYADRO/tree/Access-to-the-service)
-   [Тестирование. Покрытие и проверка гонок](https://github.com/sgsoul/golangYADRO/tree/Testing-and-verification)
-   [Web-UI. Шаблоны HTML, проверка кода и имиджей](https://github.com/sgsoul/golangYADRO/tree/Web-UI)
-   [Подключение сервиса к Telegram + gRPC аутентификация](https://github.com/sgsoul/golangYADRO/tree/gRPC-n-TGbot)

 # Описание проекта
Это сервис для поиска комиксов XKCD по запросу пользователя. Проект включает нормализацию текстовых данных, построение поискового индекса, предоставление доступа к комиксам через веб-интерфейс и API. Реализована работа с REST API, многопоточностью, кэшированием, json и MySQL базами данных, аутентификацией, авторизацией, а также тестирование с покрытием в среднем 80%, включая линтинг и проверку безопасности кода.
Проект также включает разработку Telegram бота с использованием gRPC для взаимодействия с сервисом аутентификации.

Токен бота необходимо положить в переменную среды `$TELEGRAM_BOT_TOKEN`.

Веб сервер:

![xkcd-ui](https://github.com/sgsoul/golangYADRO/assets/93263659/d7197a98-a904-41a2-a0a0-4fb9788a5e28)

Телеграм бот:

![Снимок экрана 2024-06-13 021421](https://github.com/sgsoul/golangYADRO/assets/93263659/4e69faeb-cd0c-49bb-bb83-8bf331696578)
![Снимок экрана 2024-06-13 021950](https://github.com/sgsoul/golangYADRO/assets/93263659/e791dd8d-8058-4d7f-a624-e15fabd9c929)

