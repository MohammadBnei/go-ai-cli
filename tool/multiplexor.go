package tool

type Multiplex[T any] struct {
	original <-chan T
	channels map[chan T]bool
}

func NewMultiplex[T any](original <-chan T) *Multiplex[T] {
	m := &Multiplex[T]{
		original: original,
		channels: make(map[chan T]bool),
	}

	go func(o <-chan T) {
		for v := range original {
			for k := range m.channels {
				k <- v
			}
		}
		
		for k := range m.channels {
			close(k)
		}

	}(original)

	return m
}

func (m *Multiplex[T]) Add(ch chan T) {
	m.channels[ch] = true
}

func (m *Multiplex[T]) Remove(ch chan T) {
	delete(m.channels, ch)
	close(ch)
}

func (m *Multiplex[T]) Plex() <-chan T {
	plex := make(chan T)
	m.Add(plex)

	return plex
}
