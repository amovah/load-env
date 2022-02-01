# load-env

Package load-env implements loading value from environment variables.

## Installation

Use go get

```shell
go get github.com/amovah/load-env
```

## Usage

First you need to define a struct with load-env tags:
```go
type Config struct {
    Port int `env:"name=PORT",default=8080`
    Host string `env:"name=HOST",required`
}
```

Then simply call LoadEnv on your struct:
```go
func main() {
    config := Config{}
    load.LoadEnv(&config)
    fmt.Println(config)
}
```

## Tags

- name: The name of the environment variable.
- default: The default value if the environment variable is not set.
- required: If the environment variable is not set, the function returns an error.