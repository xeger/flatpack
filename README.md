What is it?
===========

Flatpack is an experimental configuration mechanism for Go programs.

How do I use it?
----------------

Define your configuration as a plain old Go struct, potentially with nested structs to represent hierarchy. Decide on
a data source for configuration and instantiate the appropriate flatpack Getter implementation, then call Unmarshal.

```go
import (
    "github.com/xeger/flatpack"
    "github.com/xeger/flatpack/getter"
)

type Config struct {
    Database struct {
        Host string
        Port int
    }

    WidgetFactory string
}

func main() {
    config = Config{}
    env := getter.Environment{}
    err := flatpack.Unmarshal(env, config)
}
```

If Unmarshal returns no errors, your config is available and your app is ready to go!

Why?
----

Unsure; it remains to be seen whether this is useful. The goal is to:

1. Encourage the use of plain old Go data structures to specify configuration and represent it in-memory.
2. Decouple the actual reading of config data from validation, propagation across the app, data binding, and
   other concerns.
3. Transparently support various repositories for configuration data, e.g. loading a .env file during development,
   but reading keys from a Consul or Etcd agent when running in production.
