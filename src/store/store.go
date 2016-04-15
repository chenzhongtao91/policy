package store

type StoreDriver interface {
	HealthCheck() (string, error)
	Get(key string, opts map[string]string) (string, error)
	List(key string, opts map[string]string) ([]string, error)
	Set(key string, value string, opts map[string]string) error
	Remove(key string, opts map[string]string) error
	Lock() error
	Unlock() error
}

var Backend StoreDriver

func GetDriver() StoreDriver {
	return Backend
}
