# Targz

Library for packaging folders in tar.gz archives.

[Documentation on godoc.org](http://godoc.org/github.com/m90/targz)

## Installation

Installing using go get is the easiest.

    go get github.com/m90/targz

## Usage

The API is really simple, there is only one method.

* Compress

### Create an archive containing a folder

    import "github.com/m90/targz"
    ...
    err := targz.Compress("my_folder", "my_file.tar.gz")

## Contributing

All contributions are welcome! See [CONTRIBUTING](CONTRIBUTING.md) for more info.

## License

Licensed under MIT license. See [LICENSE](LICENSE) for more information.
