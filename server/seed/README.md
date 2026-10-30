# Starter data

`300sf -seed` loads these into a new database: 23 bird orders and the 542
species accepted as observed in Sweden, with Swedish, English, German and
Latin names. It adds only what's missing (matched on the Latin name), so it is
safe to run again after the files are updated.

| File | What |
|---|---|
| `orders.csv` | orders: Latin, Swedish, English, German |
| `species.csv` | species: order, Latin, Swedish, English, German |
| `generate.py` | builds both from the sources below |

## Sources

- **Which species, and their Swedish names:** [Lista över fågelarter observerade
  i Sverige (taxonomisk)](https://sv.wikipedia.org/wiki/Lista_%C3%B6ver_f%C3%A5gelarter_observerade_i_Sverige_(taxonomisk)),
  Swedish Wikipedia, following BirdLife Sverige's list. Licensed
  [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/).
- **English and German names, and orders:** Gill, F, D Donsker and P Rasmussen
  (Eds). 2026. [IOC World Bird List](https://www.worldbirdnames.org/) (v 15.2),
  multilingual edition. doi:10.14344/IOC.ML.15.2. Licensed
  [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/).
- **Order names** in Swedish, English and German are written for this project
  (`ORDERS` in `generate.py`).

So `species.csv` is shared under CC BY-SA 4.0, with the attributions above.

A few rarities have no German name in the IOC list; the app shows the English
one instead. Where BirdLife Sverige's species is wider than IOC's, `OVERRIDES`
in `generate.py` names it (*Corvus corone* includes the Hooded Crow).

## Updating

See the top of `generate.py` for where to download the two inputs, then run
`python3 -I generate.py list.wiki multiling.xlsx` here and `300sf -seed` on
each installation.
