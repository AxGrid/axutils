CHANS
=====

Подключение
```go
import "github.com/axgrid/axutils/chans"
```


Chunk Channel
-------------

```go
    chunker := NewChunkChan[int]().WithChunkSize(5).WithChunkTimeout(time.Millisecond * 100).Build()
    go func() {
        for i := 0; i < 12; i++ {
            chunker <- i
        }
    }()
    chunk := <-chunker.C()
	// [0,1,2,3,4]
    chunk := <-chunker.C()
    // [5,6,7,8,9]
    chunk := <-chunker.C()
    // [10,11]
		
```


