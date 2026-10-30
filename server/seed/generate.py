#!/usr/bin/env python3
"""Builds species.csv and orders.csv, the starter data `300sf -seed` loads.

Which species: every bird accepted as observed in Sweden, from Swedish
Wikipedia's "Lista över fågelarter observerade i Sverige (taxonomisk)"
(CC BY-SA 4.0), which follows BirdLife Sverige. Names and orders: the IOC
World Bird List's multilingual edition (CC BY 3.0), matched on the scientific
name, then on the Swedish name where the two taxonomies differ.

    python3 -I generate.py <list.wiki> <Multiling IOC.xlsx>

Get the inputs with:

    curl -o list.wiki 'https://sv.wikipedia.org/w/index.php?title=Lista_%C3%B6ver_f%C3%A5gelarter_observerade_i_Sverige_(taxonomisk)&action=raw'
    curl -o multiling.xlsx 'https://worldbirdnames.org/Multiling%20IOC%2015.2.xlsx'
"""

import csv
import os
import re
import sys
import xml.etree.ElementTree as ET
import zipfile

# Order names in the app's languages. Orders are few and their common names
# aren't in the IOC list.
ORDERS = {
    "Anseriformes": ("Andfåglar", "Waterfowl", "Gänsevögel"),
    "Galliformes": ("Hönsfåglar", "Landfowl", "Hühnervögel"),
    "Caprimulgiformes": ("Nattskärror", "Nightjars", "Nachtschwalben"),
    "Apodiformes": ("Seglare", "Swifts", "Seglervögel"),
    "Otidiformes": ("Trappfåglar", "Bustards", "Trappen"),
    "Cuculiformes": ("Gökfåglar", "Cuckoos", "Kuckucksvögel"),
    "Pterocliformes": ("Flyghöns", "Sandgrouse", "Flughühner"),
    "Columbiformes": ("Duvfåglar", "Pigeons and doves", "Taubenvögel"),
    "Gruiformes": ("Tran- och rallfåglar", "Cranes and rails", "Kranichvögel"),
    "Podicipediformes": ("Doppingar", "Grebes", "Lappentaucher"),
    "Phoenicopteriformes": ("Flamingor", "Flamingos", "Flamingos"),
    "Charadriiformes": ("Vadare, måsar och alkor", "Waders, gulls and auks", "Regenpfeiferartige"),
    "Gaviiformes": ("Lommar", "Loons", "Seetaucher"),
    "Procellariiformes": ("Stormfåglar", "Tubenoses", "Röhrennasen"),
    "Ciconiiformes": ("Storkfåglar", "Storks", "Schreitvögel"),
    "Suliformes": ("Sulfåglar", "Gannets and cormorants", "Ruderfüßer"),
    "Pelecaniformes": ("Pelikanfåglar", "Pelicans, herons and ibises", "Pelikanartige"),
    "Accipitriformes": ("Hökartade rovfåglar", "Hawks and eagles", "Greifvögel"),
    "Strigiformes": ("Ugglor", "Owls", "Eulen"),
    "Bucerotiformes": ("Härfåglar", "Hoopoes and hornbills", "Hornvogelartige"),
    "Coraciiformes": ("Praktfåglar", "Kingfishers, rollers and bee-eaters", "Rackenvögel"),
    "Piciformes": ("Hackspettartade fåglar", "Woodpeckers", "Spechtvögel"),
    "Falconiformes": ("Falkfåglar", "Falcons", "Falken"),
    "Psittaciformes": ("Papegojfåglar", "Parrots", "Papageien"),
    "Passeriformes": ("Tättingar", "Passerines", "Sperlingsvögel"),
}

# Where BirdLife Sverige's species is wider than IOC's, IOC's names fit only
# part of it.
OVERRIDES = {
    # Includes the Hooded Crow, the crow in most of Sweden.
    "Corvus corone": {"en": "Carrion and Hooded Crow", "de": "Aaskrähe"},
}

NS = "{http://schemas.openxmlformats.org/spreadsheetml/2006/main}"


def xlsx_rows(path):
    z = zipfile.ZipFile(path)
    strings = ["".join(t.text or "" for t in si.iter(NS + "t"))
               for si in ET.fromstring(z.read("xl/sharedStrings.xml")).findall(NS + "si")]
    for row in ET.fromstring(z.read("xl/worksheets/sheet1.xml")).iter(NS + "row"):
        out = {}
        for c in row.findall(NS + "c"):
            v = c.find(NS + "v")
            if v is None:
                continue
            col = re.match(r"[A-Z]+", c.get("r")).group()
            out[col] = strings[int(v.text)] if c.get("t") == "s" else v.text
        yield out


def capitalize(name):
    return name[:1].upper() + name[1:]


def main(wiki_path, xlsx_path):
    rows = xlsx_rows(xlsx_path)
    header = next(rows)
    col = {name: c for c, name in header.items()}
    ioc_by_la, ioc_by_sv = {}, {}
    for r in rows:
        bird = {
            "order": capitalize(r.get(col["Order"], "").lower()),
            "la": r.get(col[next(n for n in col if n.startswith("IOC_"))], "").strip(),
            "en": r.get(col["English"], "").strip(),
            "de": r.get(col["German"], "").strip(),
            "sv": r.get(col["Swedish"], "").strip(),
        }
        ioc_by_la[bird["la"]] = bird
        if bird["sv"]:
            ioc_by_sv[bird["sv"].lower()] = bird

    species, missing = [], []
    # A few lines have one quote too few around the scientific name.
    line = re.compile(r"^\*\s*\[\[(?:[^|\]]*\|)?([^\]]+)\]\]\s*\('{1,2}([^']+)'{1,2}\)")
    for text in open(wiki_path, encoding="utf-8"):
        m = line.match(text.strip())
        if not m:
            continue
        sv, la = m.group(1).strip(), m.group(2).strip()
        ioc = ioc_by_la.get(la) or ioc_by_sv.get(sv.lower())
        if ioc is None:
            missing.append((sv, la))
            continue
        # BirdLife Sverige's names, as the Swedish list has them.
        bird = {"order": ioc["order"], "la": la, "sv": capitalize(sv), "en": ioc["en"], "de": ioc["de"]}
        bird.update(OVERRIDES.get(la, {}))
        species.append(bird)

    unknown = sorted({s["order"] for s in species} - ORDERS.keys())
    if unknown:
        sys.exit("orders without names: " + ", ".join(unknown))

    here = os.path.dirname(os.path.abspath(__file__))
    with open(os.path.join(here, "orders.csv"), "w", newline="", encoding="utf-8") as f:
        w = csv.writer(f)
        w.writerow(["la", "sv", "en", "de"])
        for la in dict.fromkeys(s["order"] for s in species):
            w.writerow([la, *ORDERS[la]])
    with open(os.path.join(here, "species.csv"), "w", newline="", encoding="utf-8") as f:
        w = csv.DictWriter(f, fieldnames=["order", "la", "sv", "en", "de"])
        w.writeheader()
        w.writerows(species)

    print(f"{len(species)} species")
    for sv, la in missing:
        print(f"not in the IOC list, left out: {sv} ({la})", file=sys.stderr)


if __name__ == "__main__":
    main(*sys.argv[1:3])
