package pipe

type Pipe struct {
	c chan struct{}
}

func New(w int) Pipe {
	return Pipe{
		c: make(chan struct{}, w),
	}
}

// Increment will send an empty struct to the pipe's channel and block until the
// the channel is no longer full.
func (p Pipe) Increment() {
	p.c <- struct{}{}
}

// Decrement will receive an empty struct from the pipe's channel freeing space
// in the channel for more workers.
func (p Pipe) Decrement() {
	<-p.c
}

// One is a helper function. It calls Increment and then returns the Decrement
// function. Primarily used for defer calls.
func (p Pipe) One() func() {
	p.Increment()
	return p.Decrement
}
