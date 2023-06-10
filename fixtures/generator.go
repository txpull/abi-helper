package fixtures

type Generator interface {
	Generate() error
	Write() error
	Read() error
}
