# Dice roller help
To roll a set of dice, use the `/roll` command followed by one or several dice expressions.
The results will be shown in the channel by @dicerollerbot, including total results and details as needed.
Capital/small letters are interchangeable.
You can also replace `/roll` by `/analyzeroll` to see the average and probability distribution for the roll.

- **Basic dice:**
  `NdX` is `N` dice, each with `X` sides.
  `N` is assumed to be `1` if left out.
  `X` can be `%` to mean `100`.
  For example, `/roll d20` will roll a 20-sided die and `/roll 5d%` will roll five 100-sided dice.
- **Math:**
  Integers and operators `()+-*/` have their usual meanings, except `/` rounds towards zero (use double-slash `//` or division sign `÷` for real division).
  This means that you can add modifiers and roll different kinds of dice to get a total (for example `/roll 4d6+3d4+5`), or even use the dice roller as calculator (for example `/roll (5+3-2)*7/3`).
- **Keep/drop:**
  When you roll more than one die, you can add an instruction to the end to keep some and drop other dice:
  - `kM` or `khM` will keep the highest `M` dice.
  - `klM` will keep the lowest `M` dice.
  - `dM` or `dlM` will drop the lowest `M` dice.
  - `dhM` will drop the highest `M` dice.
  For example, `/roll 3d4dl1` will roll 3 4-sided dice and drop the lowest.
- **Comma separation**:
  You can provide several roll expressions in one command using commas to delimit them.
  The total for each expression will be shown, but there will be no total adding together unrelated expressions.
  Example: `/roll 1d20+4, 1d6+2`.
- **Labels:**
  After any roll expression, you can add a space and then a description.
  This can be useful to keep track of comma separated expressions.
  You can also label expressions within parentheses, causing a sum to be shown for that subexpression.
  For example, `/roll 1d20+4 to hit, (1d6+2 slashing)+(2d8 radiant) damage`.

