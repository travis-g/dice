/*
Package dice implements virtualized standard polyhedral and specialty game dice.
The dice roll calculations are intended to be cryptographically pseudo-random
through use of crypto/rand, but the entropy source used by the package is
globally configurable.

# Dice Notation

Dice notation is an algebra-like system for indicating dice rolls in games. Dice
rolls are usually given in the form AdX+B, where A is the number of X-sided dice
to roll, with an optional modifier B. B could be an integer or potentially
another dice notation string. Additionally, A can be omitted if the number of
X-sided dice to roll is 1: 1dX can be written as simply dX.
*/
package dice
