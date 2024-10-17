# axutils

`axutils` - это библиотека Go, предоставляющая различные утилиты для работы с каналами, коллекциями и другими часто используемыми структурами данных. Эта библиотека разработана для упрощения работы с асинхронными операциями и обработкой данных в Go-приложениях.

## Установка

Для установки библиотеки используйте команду:

```
go get github.com/axgrid/axutils
```

## Утилиты для работы с коллекциями

### GuavaMap

`GuavaMap` - это реализация карты с дополнительными возможностями, вдохновленная библиотекой Guava от Google.

#### Основные возможности:

- Автоматическая загрузка значений при обращении к несуществующему ключу
- Ограничение максимального количества элементов
- Автоматическая выгрузка элементов при превышении лимита
- Таймауты на чтение и запись
- Возможность блокировки для обновления значений

#### Пример использования:

```go
guavaMap := collections.NewGuavaMap[string, int]().
    WithLoadFunc(func(key string) (int, error) {
        // Загрузка значения из внешнего источника
        return len(key), nil
    }).
    WithMaxCount(100).
    WithReadTimeout(5 * time.Minute).
    Build()

value, err := guavaMap.Get("example")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Value: %d\n", value)
```

### HashSet

`HashSet` - это реализация множества на основе хэш-таблицы.

#### Основные возможности:

- Быстрое добавление и проверка наличия элементов
- Опциональное ограничение максимального количества элементов
- Потокобезопасные операции

#### Пример использования:

```go
hashSet := collections.NewHashSet[string]().
    WithMaxCount(1000).
    Build()

hashSet.Add("example")
exists := hashSet.Has("example")
fmt.Printf("'example' exists in set: %v\n", exists)
```

### MapMutex и MapRWMutex

`MapMutex` и `MapRWMutex` - это утилиты для создания отдельных блокировок для каждого ключа в map.

#### Пример использования:

```go
mapMutex := collections.NewMapMutex[string]()

mapMutex.Lock("key1")
// Критическая секция для "key1"
mapMutex.Unlock("key1")
```

### SimpleMap

`SimpleMap` - это простая реализация потокобезопасной карты.

#### Пример использования:

```go
simpleMap := collections.NewSimpleMap[string, int]()

simpleMap.Set("key", 42)
value, exists := simpleMap.Get("key")
if exists {
    fmt.Printf("Value: %d\n", value)
}
```

### ResponseMap

`ResponseMap` - это специализированная карта для работы с асинхронными ответами.

#### Основные возможности:

- Ожидание значения по ключу
- Установка значения с таймаутом
- Автоматическое удаление устаревших значений

#### Пример использования:

```go
responseMap := collections.NewResponseMap[string, int]().
    WithTimeout(5 * time.Second).
    Build()

go func() {
    time.Sleep(2 * time.Second)
    responseMap.Set("key", 42)
}()

value := responseMap.Wait("key")
fmt.Printf("Received value: %d\n", value)
```

## Заключение

Библиотека `axutils` предоставляет широкий набор инструментов для эффективной работы с данными и асинхронными операциями в Go. Использование этих утилит может значительно упростить разработку и повысить производительность ваших приложений.

Для получения дополнительной информации о других утилитах библиотеки, пожалуйста, обратитесь к документации кода или примерам использования.