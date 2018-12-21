# Dice

Dice is a CLI solution to your dice-rolling needs. It's also technically a calculator.

[Dice notation][dice-notation] is an algebra-like system for indicating dice rolls in games. Dice rolls are usually given in the form ![AdX+B](https://chart.googleapis.com/chart?cht=tx&chf=bg,s,00000000&chl=A%5Ctext%7Bd%7DX%20%5Cpm%20B), where *A* is the number of *X*-sided dice to roll, with an optional modifier *B*. *B* could be an integer or potentially another dice notation string. Additionally, _A_ can be omitted if the number of _X_-sided dice to roll is 1: 1d<em>X</em> can be written as simply d<em>X</em>.

## Build

You need [Go][golang] installed. To fetch the source and dependencies and place everything in your [GOPATH][gopath]:

```console
$ go get -u github.com/travis-g/dice/...
# downloads/updates source
```

The actual `main` package is defined in `cmd/dice`.

[dice-notation]: https://en.wikipedia.org/wiki/Dice_notation
[dice-reference]: https://wiki.roll20.net/Dice_Reference
[golang]: https://golang.org/
[gopath]: https://golang.org/doc/code.html#GOPATH
[zerolog]: https://github.com/rs/zerolog
