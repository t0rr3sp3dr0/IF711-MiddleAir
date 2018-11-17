package bonjour

type Service struct {
	UUID     string
	Provider Provider
	Tags     [12]string
	Metadata Metadata
}

type Provider struct {
	Host string
	Port uint16
}

type Metadata struct {
	OS   string
	Arch string
	Host string
	Lang string
}
