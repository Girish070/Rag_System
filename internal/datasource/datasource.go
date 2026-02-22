package datasource

type DataSource interface {
	Read() ([]byte, error)
	Metadata() map[string]string
}