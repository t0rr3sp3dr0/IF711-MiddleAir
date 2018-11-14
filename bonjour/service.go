package bonjour

type Service struct {
	Type string
	Host uint64
	Port uint16
	Call func(i interface{}) error
}
