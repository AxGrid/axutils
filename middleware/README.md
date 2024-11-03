# axmiddleware

Гибкая и типобезопасная библиотека промежуточного ПО (middleware) для Go-приложений, использующая дженерики для обработки любых типов запросов и ответов.

## Возможности

- 🔒 Типобезопасная обработка цепочки middleware с использованием дженериков
- 🛠 Паттерн Builder для удобной конфигурации
- 📝 Встроенная поддержка логирования через zerolog
- ⚡ Поддержка context.Context
- 🔄 Возможность прерывания цепочки middleware
- 🏗 Готовые реализации популярных middleware
- 💪 Встроенная защита от паник

## Установка

```bash
go get github.com/axgrid/axutils
```

## Быстрый старт

Пример использования библиотеки:

```go
package main

import (
    "context"
    "github.com/axgrid/axutils/axmiddleware"
    "github.com/rs/zerolog/log"
)

// Define request and response types
type Request struct {
    Data string
}

type Response struct {
    Result string
}

func main() {
    // Create processor
    processor := axmiddleware.NewProcessor[Request, Response]().
        WithLogger(log.Logger).
        WithMiddlewares(
            axmiddleware.CatchPanicMiddlewares[Request, Response],
            axmiddleware.LogRequestResponseMiddlewares[Request, Response],
        ).
        WithHandlers(handler).
        Build()

    // Process request
    request := Request{Data: "test"}
    response := Response{}
    
    statusCode, err := processor.Process(context.Background(), request, &response)
    if err != nil {
        log.Error().Err(err).Msg("Processing error")
    }
}

func handler(c *axmiddleware.Context[Request, Response]) {
    req := c.Request()
    c.Response().Result = "Processed: " + req.Data
}
```

## Основные концепции

### Context

Тип `Context[R, S]` предоставляет расширенный контекст для обработки, включающий:
- Доступ к запросу и ответу
- Управление значениями контекста
- Обработку ошибок
- Управление кодами статуса
- Возможности логирования
- Управление цепочкой (прерывание, продолжение)

### Цепочка Middleware

Middleware выполняются в порядке добавления. Каждый middleware может:
- Получать доступ к запросу и ответу
- Изменять контекст
- Управлять потоком выполнения
- Обрабатывать ошибки
- Устанавливать коды статуса

### Встроенные Middleware

1. **Обработка паник**
```go
axmiddleware.CatchPanicMiddlewares[R, S]
```
Перехватывает паники в цепочке middleware и логирует ошибки.

2. **Логирование запросов/ответов**
```go
axmiddleware.LogRequestResponseMiddlewares[R, S]
```
Логирует детали запросов и ответов, включая время выполнения.

## Продвинутое использование

### Создание собственного Middleware

```go
func CustomMiddleware[R, S any](ctx *axmiddleware.Context[R, S]) {
    // Pre-processing
    ctx.Logger().Info().Msg("Before processing")
    
    ctx.Next() // Continue to next middleware
    
    // Post-processing
    ctx.Logger().Info().Msg("After processing")
}
```

### Обработка ошибок

```go
func ErrorHandlingMiddleware[R, S any](ctx *axmiddleware.Context[R, S]) {
    if condition {
        ctx.AbortWithErrorAndCode(400, errors.New("invalid request"))
        return
    }
    ctx.Next()
}
```

### Работа с контекстом

```go
func ContextMiddleware[R, S any](ctx *axmiddleware.Context[R, S]) {
    ctx.WithValue("key", "value")
    ctx.Next()
    
    // Access value in next middlewares
    value := ctx.MustStringValue("key")
}
```

## Особенности использования

1. **Типобезопасность**
   - Все операции с запросами и ответами проверяются на этапе компиляции
   - Исключены ошибки приведения типов во время выполнения

2. **Производительность**
   - Минимальные накладные расходы благодаря использованию дженериков
   - Эффективная обработка цепочки middleware

3. **Расширяемость**
   - Легко добавлять новые middleware
   - Возможность создания составных middleware
   - Гибкая настройка под различные сценарии использования

## Требования

- Go 1.18 или выше (необходима поддержка дженериков)
- github.com/rs/zerolog для логирования

## Лицензия

MIT

## Участие в разработке

Мы приветствуем ваше участие в развитии проекта! Создавайте issue и присылайте pull request'ы.