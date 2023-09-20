#!/usr/bin/env bash
# ./json2ts.sh en sv

echo "class Translations {"

for arg in "$@"; do
	echo "	static $arg = {"

	key_value_pairs=()
	while IFS="=" read -r key value; do
		key_value_pairs+=("$key: \"$value\"")
	done < <(jq -r 'to_entries[] | "\(.key)=\(.value)"' "./locale/$arg.json")


	for pair in "${key_value_pairs[@]}"; do
		echo "		$pair,"
	done

	echo "	}"
done

echo "}"
