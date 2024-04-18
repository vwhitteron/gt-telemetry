package telemetrysrc

type Reader interface {
	Read() (int, []byte, error)
	Close() error
}
