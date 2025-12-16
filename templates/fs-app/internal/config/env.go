package config

type Environment string

const (
	Prod  Environment = "prod"
	Dev   Environment = "dev"
	Local Environment = "local"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) Prod() bool {
	return e == Prod
}

func (e Environment) Dev() bool {
	return e == Dev
}

func (e Environment) Local() bool {
	return e == Local
}
