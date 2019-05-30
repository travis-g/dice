/*
Package dice implements virtualized standard polyhedral and specialty game dice.
The dice roll calculations are intended to be cryptographically pseudo-random
through use of crypto/rand, but the entropy source used by the package is
globally configurable.
*/
package dice
