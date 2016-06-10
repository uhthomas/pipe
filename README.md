# Pipe
A simple package which enables the piping of functions without the need for worker pools. Inspired by personal needs and [Marcio's approach to concurrency patterns.](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/)

## Usage
`Increment` and `Decrement` **must** always be called otherwise the pipe will become out of sync.

```go
package main

import (
  "fmt"
  "time"

  "github.com/6f7262/pipe"
)

func main() {
  // Create a new pipe with a maximum of 10 workers
  p := pipe.New(10)

  for i := 0; i < 100; i++ {
    go worker(p, i) 
  }

  select{}
}

func worker(p pipe.Pipe, i int) {
    defer p.One()()
    
    // Hard work...
    time.Sleep(1 * time.Second)
    fmt.Printf("Worker %d finished\n", i)
}
```

## Why would I use this?
Pipe is designed to help ease the pain of limiting the amount of work done at one time. A normal way to approach this would be to create a worker pool such as
```go
package main

import (
  "fmt"
  "time"
)

type Pool struct {
  C chan int
}

func NewPool(w int) Pool {
  p := Pool{C: make(chan int, w)}
  for i := 0; i < w; i++ {
    go p.worker()
  }
  return p
}

func (p Pool) worker() {
  for {
    i := <-p.C
    
    // Hard work
    time.Sleep(1 * time.Second)
    fmt.Printf("Worker %d finished\n", i)
  }
}

func main() {
  p := NewPool(10)
  for i := 0; i < 100; i++ {
    p.C <- i
  }

  select{}
}
```
The code above does exactly the same work as the example code for Pipe but instead of being 2-3 extra lines of code, it's maybe 10 or more depending on your implementation (Marcio's was almost 100).

The primary advantage of this over a worker pool is that you can return values from a function whereas in a worker pool you can't. Pipe also spawns far less goroutines and therefore saves on memory and compute power as it doesn't always have 10 or more workers running in the background, the Pipe is just a struct with an single channel.

## Patterns
There are some patterns which can come in handy which some other worker pool implementations use. We're going to look at this from the perspective of receiving 100's of HTTP requests.

### Not waiting for the work to finish
In Marcio's example the job queue would only block until the work is received, once it was received it would unblock and the request would proceed as normal, even if the upload had failed. It's quite easy to do that with the Pipe package, instead of defering `Decrement` in the worker, defer it in a new goroutine such as so.
```go
func main() {
    p := pipe.New(10)
    for i := 0; i < 180; i++ {
        worker(p, i)
    }
}

func worker(p pipe.Pipe, v interface{}) {
    p.Increment()
    go func() {
        defer p.Decrement()
        // Hard work...
    }()
}
```

### Payload and worker options
Some people may want to have a large buffer for the amount of work they want in a pipe while still only having maybe 10 workers actually doing that work. A simple solution for this would be to make a channel in which work is sent to and a router which sends the work from that channel to the worker.
```go
func main() {
    ch := make(chan interface{}, 200)
    go func() {
        p := pipe.New(10)
        for v := range ch {
            p.Increment()
            go worker(p, v)
            // Alternatively, the worker function can be called synchronously
            // and follow the "Not waiting for work to finish" pattern.
            // This method just demonstrates that pipes can be passed to other
            // functions and the worker doesn't always have to be the one to
            // call Increment.
        } 
    }()

    for i := 0; i < 180; i++ {
        ch <- i
    }
}

func worker(p pipe.Pipe, v interface{}) {
    defer p.Decrement()
    // Hard work
}
```

