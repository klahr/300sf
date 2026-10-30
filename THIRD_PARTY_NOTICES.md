# Third-party notices

300sf, including its artwork (logo, icons and illustrations), is
Copyright © 2022–2026 Angels' Ware and licensed under the AGPL-3.0-or-later
(see [`LICENSE`](LICENSE)). It ships with the following works by others, each
under its own license.

## In the web pages (`server/site/static/`)

| Work | Files | License |
|---|---|---|
| [jQuery](https://jquery.com) 3.7.1, © OpenJS Foundation and other contributors | `vendor/jquery/` | MIT, below |
| [Leaflet](https://leafletjs.com) 1.9.4, © 2010–2023 Vladimir Agafonkin, © 2010–2011 CloudMade | `vendor/leaflet/` | BSD-2-Clause, below |
| [Swagger UI](https://github.com/swagger-api/swagger-ui) 5, © SmartBear Software | `vendor/swagger-ui/` | [Apache-2.0](LICENSES/Apache-2.0.txt) |
| [Inter](https://rsms.me/inter/), © 2016 The Inter Project Authors | `fonts/inter-*.woff2` | [OFL-1.1](LICENSES/OFL-1.1.txt) |
| [Source Serif 4](https://github.com/adobe-fonts/source-serif), © 2014–2021 Adobe | `fonts/source-serif-4-*.woff2` | [OFL-1.1](LICENSES/OFL-1.1.txt) |
| [Material Symbols](https://github.com/google/material-design-icons), © Google | `fonts/material-symbols-rounded.woff2` | [Apache-2.0](LICENSES/Apache-2.0.txt) |

## Data

| Work | Where | License |
|---|---|---|
| [IOC World Bird List](https://www.worldbirdnames.org/) v15.2 (Gill, Donsker & Rasmussen) | `server/seed/` | CC BY 3.0 |
| [Swedish Wikipedia](https://sv.wikipedia.org/wiki/Lista_%C3%B6ver_f%C3%A5gelarter_observerade_i_Sverige_(taxonomisk)): birds observed in Sweden | `server/seed/species.csv` | CC BY-SA 4.0 |
| [OpenStreetMap](https://www.openstreetmap.org/copyright) map tiles, and the places zones are made from at run time | fetched, not shipped | ODbL 1.0 |

See [`server/seed/README.md`](server/seed/README.md) for how the starter data
is made.

## Services the code uses

- **BirdNET** ([BirdNET-Analyzer](https://github.com/kahst/BirdNET-Analyzer)),
  in `analyzer/`: its model is licensed CC BY-NC-SA 4.0, for non-commercial
  use only.
- **xeno-canto** recordings and **Wikipedia/Wikimedia Commons** articles and
  photos are loaded from those services under their own licenses, shown with
  each.

The Go and npm dependencies are listed in `server/go.mod` and
`frontend/package.json` with their licenses in their own repositories.

## MIT (jQuery)

```
Copyright OpenJS Foundation and other contributors, https://openjsf.org/

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
```

## BSD-2-Clause (Leaflet)

```
Copyright (c) 2010-2023, Vladimir Agafonkin
Copyright (c) 2010-2011, CloudMade
All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are
permitted provided that the following conditions are met:

   1. Redistributions of source code must retain the above copyright notice, this list of
      conditions and the following disclaimer.

   2. Redistributions in binary form must reproduce the above copyright notice, this list
      of conditions and the following disclaimer in the documentation and/or other materials
      provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
THE COPYRIGHT HOLDERS OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT
OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```
