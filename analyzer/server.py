"""Bird sound identification for the 300sf server.

POST /analyze with a multipart "audio" file and an optional "meta" JSON field
({"num_results": 5, "sensitivity": 1, "week": 1-48, "lat": .., "lon": ..}).
Answers {"msg": "success", "results": [["Latin name_Common name", confidence], ...]},
best first, one entry per species. BirdNET-Analyzer 2.4 no longer has a server of its own.
"""

import csv
import glob
import json
import os
import tempfile
import threading
import wave

import bottle

import birdnet_analyzer

# The centre of Sweden; with the week this limits the answers to birds that can be here now.
DEFAULT_LAT = 62.0
DEFAULT_LON = 16.0

# BirdNET keeps its settings in module globals, so one analysis at a time.
lock = threading.Lock()


def number(meta, key, default, low, high):
    try:
        value = float(meta.get(key, default))
    except (TypeError, ValueError):
        return default
    return min(high, max(low, value))


@bottle.post("/analyze")
def analyze():
    upload = bottle.request.files.get("audio")
    if upload is None:
        bottle.response.status = 400
        return {"msg": "missing audio"}
    try:
        meta = json.loads(bottle.request.forms.get("meta") or "{}")
    except ValueError:
        meta = {}
    if not isinstance(meta, dict):
        meta = {}

    num_results = int(number(meta, "num_results", 5, 1, 20))
    sensitivity = number(meta, "sensitivity", 1.0, 0.5, 1.5)
    week = int(number(meta, "week", -1, -1, 48))
    lat = number(meta, "lat", DEFAULT_LAT, -90, 90)
    lon = number(meta, "lon", DEFAULT_LON, -180, 180)

    with tempfile.TemporaryDirectory() as work:
        path = os.path.join(work, "recording" + os.path.splitext(upload.raw_filename or "")[1][:8].lower())
        upload.save(path)
        out = os.path.join(work, "out")
        os.mkdir(out)
        with lock:
            birdnet_analyzer.analyze(
                path, output=out, min_conf=0.1, lat=lat, lon=lon, week=week,
                sensitivity=sensitivity, rtype="csv", threads=1,
            )

        best = {}
        for result in glob.glob(os.path.join(out, "*.BirdNET.results.csv")):
            with open(result, newline="", encoding="utf-8") as f:
                for row in csv.DictReader(f):
                    label = row["Scientific name"] + "_" + row["Common name"]
                    best[label] = max(best.get(label, 0.0), float(row["Confidence"]))

    results = sorted(best.items(), key=lambda e: e[1], reverse=True)[:num_results]
    return {"msg": "success", "results": [[label, round(conf, 4)] for label, conf in results]}


def warm_up():
    # Loading the model takes a while; do it before the first user waits for it.
    with tempfile.TemporaryDirectory() as work:
        path = os.path.join(work, "silence.wav")
        with wave.open(path, "wb") as w:
            w.setnchannels(1)
            w.setsampwidth(2)
            w.setframerate(48000)
            w.writeframes(b"\0\0" * 48000 * 3)
        birdnet_analyzer.analyze(path, output=work, rtype="csv", threads=1)


if __name__ == "__main__":
    warm_up()
    bottle.BaseRequest.MEMFILE_MAX = 10 * 1024 * 1024
    bottle.run(host="0.0.0.0", port=int(os.environ.get("PORT", "8080")), quiet=True)
