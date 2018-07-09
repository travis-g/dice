# DRAAS

DRAAS (Dice Rolls As A Service) is a scalable HTTP API solution to your dice-rolling needs. It's also technically a calculator.

[Dice notation][dice-notation] is an algebra-like system for indicating dice rolls in games. Dice rolls are usually given in the form **<em>X</em>d<em>Y</em>((-|+)<em>N</em>)**, where *X* is the number of *Y*-sided dice to roll, with an optional modifier *N*. *N* could be an integer or potentially another dice notation string.

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

## Todo

- [ ] `MaxDice` enforcement to deny arbitrarily large and taxing dice rolls.
- [ ] Better request and response logging.
- [ ] Function support, such as `floor()`, `abs()`, `max()`, `min()`, etc.
- [ ] Drop/keep support, such as `5d6d3` to drop the lowest 3 results, or `5d6k3` to keep the highest 3.
- [ ] Advantage/disadvantage support, such as `adv(d20)`. Similar functionality could be achieved with drop/keep via `2d20k1` to roll twice and keep the highest.

[dice-notation]: https://en.wikipedia.org/wiki/Dice_notation
[dice-reference]: https://wiki.roll20.net/Dice_Reference
