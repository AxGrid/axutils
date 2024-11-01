# axutils

`axutils` - это библиотека Go, предоставляющая различные утилиты для работы с каналами, коллекциями и другими часто используемыми структурами данных. Эта библиотека разработана для упрощения работы с асинхронными операциями и обработкой данных в Go-приложениях.

## Установка

Для установки библиотеки используйте команду:

```
go get github.com/axgrid/axutils
```

## Утилиты для работы с каналами

### ChunkChan

`ChunkChan` - это структура, которая группирует элементы из входящего канала в чанки (группы) заданного размера или по истечении заданного времени.

#### Основные возможности:

- Группировка элементов в чанки заданного размера
- Отправка чанков по таймауту, если размер чанка не достигнут
- Настраиваемые размеры буферов для входящего и исходящего каналов
- Возможность задать функцию для обработки чанков

#### Пример использования:

```go
chunkChan := chans.NewChunkChan[int]().
    WithChunkSize(100).
    WithChunkTimeout(50 * time.Millisecond).
    Build()

// Добавление элементов
go func() {
    for i := 0; i < 1000; i++ {
        chunkChan.Add(i)
    }
}()

// Получение и обработка чанков
for chunk := range chunkChan.C() {
    fmt.Printf("Received chunk of size %d: %v\n", len(chunk), chunk)
}
```

В этом примере, `chunk` будет срезом целых чисел (`[]int`). Каждый чанк будет содержать до 100 элементов или меньше, если прошло 50 миллисекунд с момента получения последнего элемента.

### ShardChan

`ShardChan` - это структура, которая распределяет входящие элементы по нескольким исходящим каналам (шардам) на основе функции шардирования.

#### Основные возможности:

- Распределение элементов по шардам с помощью пользовательской функции
- Настраиваемое количество шардов
- Возможность задать функцию-обработчик для каждого шарда

#### Пример использования:

```go
shardChan, _ := chans.NewShardChan[string]().
    WithShardCount(4).
    WithShardFunc(func(s string) int {
        return len(s)
    }).
    Build()

// Добавление элементов
go func() {
    words := []string{"Hello", "World", "Go", "Programming"}
    for _, word := range words {
        shardChan.Add(word)
    }
}()

// Получение и обработка элементов из каждого шарда
for i := 0; i < shardChan.ShardCount(); i++ {
    go func(shardIndex int) {
        for item := range shardChan.C(shardIndex) {
            fmt.Printf("Shard %d received: %s\n", shardIndex, item)
        }
    }(i)
}
```

В этом примере, элементы (строки) будут распределены по 4 шардам в зависимости от их длины. Каждый шард (`shardChan.C(shardIndex)`) будет содержать отдельный поток элементов.

### ShardChunk

`ShardChunk` - это комбинация `ShardChan` и `ChunkChan`, которая сначала распределяет элементы по шардам, а затем группирует их в чанки внутри каждого шарда.

#### Основные возможности:

- Распределение элементов по шардам
- Группировка элементов в чанки внутри каждого шарда
- Настраиваемые параметры шардирования и группировки
- Возможность задать функцию-обработчик для чанков в каждом шарде

#### Пример использования:

```go
shardChunk, _ := chans.NewShardChunk[int]().
    WithShardCount(4).
    WithChunkSize(10).
    WithChunkTimeout(50 * time.Millisecond).
    WithShardFunc(func(i int) int {
        return i % 4
    }).
    Build()

// Добавление элементов
go func() {
    for i := 0; i < 100; i++ {
        shardChunk.Add(i)
    }
}()

// Получение и обработка чанков из каждого шарда
for i := 0; i < shardChunk.ShardCount(); i++ {
    go func(shardIndex int) {
        for chunk := range shardChunk.C(shardIndex) {
            fmt.Printf("Shard %d received chunk: %v\n", shardIndex, chunk)
        }
    }(i)
}
```

В этом примере, числа будут распределены по 4 шардам в зависимости от остатка от деления на 4. Внутри каждого шарда числа будут сгруппированы в чанки по 10 элементов или меньше, если прошло 50 миллисекунд. Каждый шард (`shardChunk.C(shardIndex)`) будет выдавать чанки в виде `[]int`.

## Заключение

Библиотека `axutils` предоставляет мощные инструменты для работы с асинхронными операциями и обработкой данных в Go. Использование `ChunkChan`, `ShardChan` и `ShardChunk` позволяет эффективно управлять потоками данных, распределять нагрузку и группировать элементы для дальнейшей обработки.

При работе с этими утилитами важно помнить:

1. `ChunkChan` возвращает срезы элементов (`[]T`), где `T` - тип данных, с которым вы работаете.
2. `ShardChan` распределяет отдельные элементы по нескольким каналам, каждый из которых может быть обработан независимо.
3. `ShardChunk` комбинирует оба подхода, предоставляя срезы элементов (`[]T`) для каждого шарда.

Эти структуры данных особенно полезны при обработке больших объемов данных, балансировке нагрузки между несколькими обработчиками или при необходимости группировки связанных элементов для пакетной обработки.

Для получения дополнительной информации о других утилитах библиотеки, пожалуйста, обратитесь к документации кода или примерам использования.