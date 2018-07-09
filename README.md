# DRAAS

DRAAS (Dice Rolls As A Service) is a scalable HTTP API solution to your dice-rolling needs. It's also technically a calculator.

[Dice notation][dice-notation] is an algebra-like system for indicating dice rolls in games. Dice rolls are usually given in the form **<em>X</em>d<em>Y</em>((-|+)<em>N</em>)**, where *X* is the number of *Y*-sided dice to roll, with an optional modifier *N*. *N* could be an integer or potentially another dice notation string. Additionally, _X_ can be omitted if the number of _Y_-sided dice to roll is 1: 1d<em>Y</em> can be written as simply d<em>Y</em>.

## Usage

Full DRAAS options are available with the `-h` flag. To start a DRAAS server listening on the default port:

```console
$ ./draas
{"level":"info","time":"2018-07-09T11:47:24-04:00","message":"server started"}
...
```

The server will gracefully shut down if `SIGINT`/<kdb>^C</kbd> is sent to the process. DRAAS also supports a `-debug` flag to increase verbosity.

The `-pretty` flag will set DRAAS to output server logs with [zerolog][zerolog]'s pretty logging:

```console
$ draas -debug
2018-07-09T11:50:39-04:00 |DEBU| debug mode enabled
2018-07-09T11:50:39-04:00 |DEBU| seeded PRNG seed=1531151439237556569
2018-07-09T11:50:39-04:00 |INFO| server started
2018-07-09T11:51:04-04:00 |DEBU| rolled expanded=5+1 result=6 roll=2d6
2018-07-09T11:51:10-04:00 |DEBU| rolled expanded=5 result=5 roll=1d20
2018-07-09T11:51:21-04:00 |INFO| SIGINT received
2018-07-09T11:51:21-04:00 |INFO| shutting down
```

## API

### Evaluate a Roll Expression

This endpoint returns the result of evaluating given [dice notation][dice-notation] expressions.

| Method | Path            | Produces               |
| ------ | --------------- | ---------------------- |
| `GET`  | `/(roll/):dice` | `200 application/json` |

If you include spaces in your roll (ex. `3d6 + 1`) the request will need to be URL encoded. Non-encoded spaces will result in a `400 bad request`.

#### Parameters

- `dice (string: <required>)` - Specifies the [dice notation](#dice-notation) expression. The expression should have at least one dice notation substring.

#### Sample Request

```console
$ curl \
    --request GET \
    localhost:8000/1d20+1
```

#### Sample Response

```json
{
  "expanded": "(8)+1",
  "result": 9
}
```

## HTTP Status Codes

- `200` - Success with roll result
- `400` - Invalid request, usually due to a dice notation syntax error.
- `404` - Invalid path. This can mean that the requested resource did not exist or that the dice notation was not interpreted correctly.
- `414` - Request-URI was too long or requested too many dice rolls.
- `500` - Internal server error. Try again later, and contact the maintainer if the problem persists.

## Notes

- The dice rolls are pseudo-random. `crypto/rand` would be an easy swap-in, but a full CSPRNG integration is significantly slower compared to traditional seeding of `math.rand` with the system time. My intention is to implement `crypto/rand` in some capacity later.

## Build

Compile DRAAS with `go build`. If dependencies are missing from [GOPATH][gopath] they can be all be fetched at once with `go get ./...`.

## Todo

- [ ] `MaxDice` enforcement to deny arbitrarily large and taxing dice rolls.
- [ ] Better request and response logging.
- [ ] Function support, such as `floor()`, `abs()`, `max()`, `min()`, etc.
- [ ] Drop/keep support, such as `5d6d3` to drop the lowest 3 results, or `5d6k3` to keep the highest 3.
- [ ] Advantage/disadvantage support, such as `adv(d20)`. Similar functionality could be achieved with drop/keep via `2d20k1` to roll twice and keep the highest.

[dice-notation]: https://en.wikipedia.org/wiki/Dice_notation
[dice-reference]: https://wiki.roll20.net/Dice_Reference
[gopath]: https://golang.org/doc/code.html#GOPATH
[zerolog]: https://github.com/rs/zerolog
