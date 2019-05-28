# Dice

[![GoDoc](https://godoc.org/github.com/travis-g/dice?status.svg)](https://godoc.org/github.com/travis-g/dice) [![Go Report Card](https://goreportcard.com/badge/github.com/travis-g/dice)](https://goreportcard.com/report/github.com/travis-g/dice)

Dice is a CLI solution to your dice-rolling needs. It's also technically a calculator.

[Dice notation][dice-notation] is an algebra-like system for indicating dice rolls in games. Dice rolls are usually given in the form ![AdX+B][tex-AdX+B], where ![A][tex-A] is the number of ![X][tex-X]-sided dice to roll, with an optional modifier ![B][tex-B]. ![B][tex-B] could be an integer or potentially another dice notation string. Additionally, ![A][tex-A] can be omitted if the number of ![X][tex-X]-sided dice to roll is 1: ![1dX][tex-1dX] can be written as simply ![dX][tex-dX].

## Install

You need [Go][golang] installed. To fetch just the main CLI and place it in your [GOPATH][gopath]:

```console
go get -u github.com/travis-g/dice/cmd/dice
```

## Build

To fetch the source and dependencies and place everything in your [GOPATH][gopath]:

```console
go get -u github.com/travis-g/dice/...
```

The actual `main` package is defined in `cmd/dice`. To test everything and build the CLI:

```console
make build
```

See the `Makefile` for more.

## Tips

- Alias `dice eval` as `roll` in your shell if you get sick of specifying the subcommand.

  ```sh
  alias roll="dice eval"
  ```

[dice-notation]: https://en.wikipedia.org/wiki/Dice_notation
[dice-reference]: https://wiki.roll20.net/Dice_Reference
[golang]: https://golang.org/
[gopath]: https://golang.org/doc/code.html#GOPATH

[tex-1dX]: https://chart.googleapis.com/chart?cht=tx&chf=bg,s,00000000&chl=%5Ctext%7B1d%7DX
[tex-A]: https://chart.googleapis.com/chart?cht=tx&chf=bg,s,00000000&chl=A
[tex-AdX+B]: https://chart.googleapis.com/chart?cht=tx&chf=bg,s,00000000&chl=A%5Ctext%7Bd%7DX%20%5Cpm%20B
[tex-B]: https://chart.googleapis.com/chart?cht=tx&chf=bg,s,00000000&chl=B
[tex-dX]: https://chart.googleapis.com/chart?cht=tx&chf=bg,s,00000000&chl=%5Ctext%7Bd%7DX
[tex-X]: https://chart.googleapis.com/chart?cht=tx&chf=bg,s,00000000&chl=X
