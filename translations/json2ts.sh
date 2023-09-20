#!/usr/bin/env bash
# ./json2ts.sh en sv

echo "class Translations {"

for arg in "$@"; do
    echo "    static $arg = {"

    mapfile -t key_value_pairs < <(jq -j 'to_entries[] | "\(.key): \"\( .value | gsub("\\n"; "<br/>") | gsub("\\r"; "") )\"\n"' "./locale/$arg.json")

    for pair in "${key_value_pairs[@]}"; do
        echo "        $pair,"
    done

    echo "    }"
done

echo "}"
