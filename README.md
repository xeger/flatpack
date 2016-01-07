What is it?
===========

Flatpack is a [12-factor style](http://12factor.net/config) configuration mechanism
for Go programs. It reads data from the process environment and "splats" it into a struct of
your choice. You benefit from Go's type safety without writing boilerplate config-loading code;
your users never need to touch a config file; you get to spend your time on features that _matter_.

![Build Status](https://travis-ci.org/xeger/flatpack.svg) [![Coverage Status](https://coveralls.io/repos/xeger/flatpack/badge.svg?branch=master&service=github)](https://coveralls.io/github/xeger/flatpack?branch=master)

How do I use it?
----------------

Populate the environment with some inputs that you want to convey to your app.

```bash
export DATABASE_HOST=db1.example.com
export DATABASE_PORT=1234
export WIDGET_FACTORY=jumbo
```

Next, define your configuration as a plain old Go struct, potentially with nested structs to represent hierarchy.
Ask flatpack to unmarshal the environment into your data structure.

```go
import (
    "github.com/xeger/flatpack"
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
    err := flatpack.Unmarshal(&config)
}
```

If Unmarshal returns no errors, your config is available and your app is ready to go!

Why should I use it?
----

Flatpack's goals are to:

1. Encourage the use of plain old Go data structures to specify configuration and represent it in-memory.
2. Decouple the actual reading of config data from validation, propagation across the app, data binding, and
   other concerns.
3. Transparently support various repositories for configuration data, e.g. loading a .env file during development,
   but reading keys from a Consul or Etcd agent when running in production.

Details
=======

Flatpack uses the `reflect` package to recurse through your configuration struct.
As it goes, it maintains a list of field names that it used to arrive at its present
location. This list of field names is transformed into an underscore-delimited string
like so:
 * `Foo.Bar` becomes `FOO_BAR_BAZ`
 * `Foo.Bar.BazQuux` becomes `FOO_BAR_BAZ_QUUX`

(Yes, this means that `Foo.BarBaz` and `FooBar.Baz` will both be populated from the same
environment variable; don't do that! In the future, flatpack will count this as an
error and refuse to load your struct.)

If the environment variable is defined, flatpack parses its value and coerces it to
the data type of that field. Supported data types are booleans, numbers, strings,
and lists of any of those. If a coercion fails, flatpack returns an error and your
app exits with a useful message about what's wrong in the config.

As a _coup de grace_, flatpack calls `Validate()` on your configuration object
if it defines that method, giving you a chance to validate the finer points of
your configuration or log a startup message with config details.

What Next?
----------

The `flatpack.Getter` interface allows us to load data from sources other than
the environment; it might make sense to build support for HTTP key/value stores
directly into the library.

On the other hand, I like the idea of using the process environment to decouple
the producer of config data from the consumer; it produces a naturally-portable
app. I might find or create a shim that reads key/value stores and writes to
the environment.
