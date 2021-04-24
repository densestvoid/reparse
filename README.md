# Version 1.0 Release Date: TBD
This package is only tested from an academic perspective. Once it has been successfully used and tested in the wild, I will consider releasing a version

<br>

# StructExp
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/densestvoid/structexp?label=version&logo=version&sort=semver)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/densestvoid/structexp)](https://pkg.go.dev/github.com/densestvoid/structexp)

### A package to parse structs from strings using the power of regular expression

## Example
```golang
package main

import (
	"fmt"

	"github.com/densestvoid/structexp"
)

type PhoneNumber struct {
    structexp.StructExp `structexp:"{{cc}}-{{ac}}-{{p}}-{{ln}}"`
    CountryCode int `structexp.name:"cc"`
    AreaCode int `structexp.name:"ac"`
    Prefix int `structexp.name:"p"`
    LineNumber int `structexp.name:"ln"`
}

func main() {
    var number PhoneNumber
    if err := structexp.Parse("1-234-567-8900", &number); err != nil {
        fmt.Println(err)
    }
    fmt.Println(number)
}
```

## Installation
`go get github.com/densestvoid/structexp`

## Support
Join the [Discord server](https://discord.gg/raAdxWuKTU).

## Contribute
Feel free to tackle any open issues, or if a feature request catches your eye, feel free to reach out to me and we can discuss adding it to the package. If you have any ideas on expanding and adding a feature, please message me or open an issue on GitHub
## License

GPL-3.0 License © [DensestVoid](https://github.com/densestvoid)