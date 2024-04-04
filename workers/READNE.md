WORKERS
=======


Подключить
```go

import "github.com/axgrid/axutils/workers"

```


Runner
------

Создает какое-то количество воркеров и запускает их.

```go

func main() {
    r := NewRunner().
        WithWorkerCount(3). // Количество воркеров
        Build()
    // Запускаем тяжелую задачу        
    r.Run(func() {
        time.Sleep(10 * time.Millisecond)
    })
}

```

