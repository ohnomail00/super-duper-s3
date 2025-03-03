package clients

type Factory interface {
	New(addr string) Client
}

type DummyFactory struct{}

func NewDummyFactory() Factory {
	return &DummyFactory{}
}

func (f *DummyFactory) New(addr string) Client {
	return NewStorage(addr)
}
