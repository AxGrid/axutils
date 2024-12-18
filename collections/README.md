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

### WaitMap

`WaitMap` - это реализация карты с поддержкой ожидания значений и автоматической очисткой устаревших данных.

#### Основные возможности:

- Асинхронное ожидание значений
- Автоматическая очистка данных по TTL
- Таймауты на запросы
- Контекст для управления жизненным циклом
- Потокобезопасные операции

#### Пример использования:

```go
waitMap := collections.NewWaitMap[string, int]().
    WithRequestTimeout(10 * time.Second).
    WithResponseTtl(5 * time.Minute).
    WithContext(ctx).
    Build()

// Асинхронная установка значения
go func() {
    time.Sleep(2 * time.Second)
    waitMap.Set("key", 42)
}()

// Ожидание значения
value := waitMap.Wait("key")
fmt.Printf("Received value: %d\n", value)
```

### StructUniqSet

`StructUniqSet` - это реализация множества для хранения уникальных структур с поддержкой TTL.

#### Основные возможности:

- Хранение уникальных структур любого типа
- Автоматическая очистка устаревших элементов
- Настраиваемая функция построения хеша
- Потокобезопасные операции
- Контроль времени жизни элементов

#### Пример использования:

```go
type MyStruct struct {
    ID   int
    Name string
}

set := collections.NewStructUniqSet[MyStruct]().
    WithElementTtl(30 * time.Minute).
    WithCheckInterval(time.Second).
    WithContext(ctx).
    Build()

item := MyStruct{ID: 1, Name: "example"}
added, err := set.Add(item)
if err != nil {
    log.Fatal(err)
}

exists, _ := set.Has(item)
fmt.Printf("Item exists in set: %v\n", exists)
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

### RequestMap

`RequestMap` - это потокобезопасная реализация карты для кэширования асинхронных запросов с автоматической дедупликацией и очисткой устаревших данных.

#### Основные возможности:

- Автоматическая дедупликация параллельных запросов с одинаковым ключом
- Кэширование результатов с настраиваемым временем жизни (TTL)
- Автоматическая очистка устаревших данных
- Потокобезопасные операции
- Поддержка обобщённых типов для ключей и значений
- Обработка ошибок при загрузке данных
- Поддержка таймаутов для операций загрузки

#### Методы:

- `GetOrCreate(key K, f func(k K) V) V` - получает значение по ключу или создает новое, используя переданную функцию
- `GetOrCreateWithErr(key K, f func(k K) (V, error)) (V, error)` - аналогичен `GetOrCreate`, но поддерживает обработку ошибок
- `Timeout(duration time.Duration, f func(k K) (V, error)) func(k K) (V, error)` - создает функцию с таймаутом для загрузки данных

#### Примеры использования:

```go
// Создаем карту с временем жизни кэша 5 минут
requestMap := collections.NewRequestMap[string, int](5 * time.Minute)

// Базовый пример с GetOrCreate
result := requestMap.GetOrCreate("example", func(key string) int {
    // Имитация загрузки данных
    time.Sleep(time.Second)
    return len(key)
})
fmt.Printf("Simple result: %d\n", result)

// Пример с обработкой ошибок
result, err := requestMap.GetOrCreateWithErr("example", func(key string) (int, error) {
    // Имитация загрузки данных с возможной ошибкой
    if len(key) == 0 {
        return 0, errors.New("empty key")
    }
    return len(key), nil
})
if err != nil {
    log.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Result with error handling: %d\n", result)
}

// Пример использования таймаута
loadWithTimeout := requestMap.Timeout(2*time.Second, func(key string) (int, error) {
    // Имитация долгой загрузки
    time.Sleep(3 * time.Second)
    return len(key), nil
})

result, err = loadWithTimeout("example")
if err == collections.ErrTimeout {
    fmt.Println("Operation timed out")
} else if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Result with timeout: %d\n", result)
}
```

## Заключение

Библиотека `axutils` предоставляет широкий набор инструментов для эффективной работы с данными и асинхронными операциями в Go. Использование этих утилит может значительно упростить разработку и повысить производительность ваших приложений.

Для получения дополнительной информации о других утилитах библиотеки, пожалуйста, обратитесь к документации кода или примерам использования.