#!/bin/bash

set -e

# Check if a file path is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <file_path>"
    exit 1
fi

# Get the directory where the script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( dirname "$SCRIPT_DIR" )"

INPUT_FILE=$1
BASENAME=$(basename "$INPUT_FILE")
EXTENSION="${BASENAME##*.}"
FILENAME="${BASENAME%.*}"

# Temporary files
JSON_FILE="/tmp/${FILENAME}_tmp.json"
YAML_FILE="/tmp/${FILENAME}_tmp.yaml"
CSV_FILE="/tmp/${FILENAME}_tmp.csv"

# Change to project root to ensure pfutil is found
cd "$PROJECT_ROOT"

# Function to compare files
compare_files() {
    if ! diff -q "$1" "$2" >/dev/null 2>&1; then
        echo "Error: Files $1 and $2 are different"
        echo "Differences:"
        diff -u "$1" "$2"
        return 1
    fi
    return 0
}

# Function to convert file
convert_file() {
    local input_file=$1
    local output_file=$2
    local output_format=$(basename "$output_file" | sed 's/.*\.//')
    
    if [[ "$output_format" == "2" ]]; then
        output_format="json"
    fi

    if "$PROJECT_ROOT/pfutil" convert "$input_file" "$output_file" -o "$output_format"; then
        echo "Successfully converted $input_file to $output_file"
    else
        echo "Failed to convert $input_file to $output_file"
        return 1
    fi
}

# Convert input file to JSON
echo "Converting $INPUT_FILE to JSON..."
if ! convert_file "$INPUT_FILE" "$JSON_FILE"; then
    echo "Skipping unsupported file type: $INPUT_FILE"
    exit 0
fi

# Convert JSON to YAML
echo "Converting JSON to YAML..."
convert_file "$JSON_FILE" "$YAML_FILE"

# Convert YAML back to JSON
echo "Converting YAML back to JSON..."
convert_file "$YAML_FILE" "${JSON_FILE}.2"

# Compare original JSON with the round-trip JSON
echo "Comparing JSON files..."
if ! compare_files "$JSON_FILE" "${JSON_FILE}.2"; then
    echo "JSON round-trip failed"
    exit 1
fi

# If the input file was YAML or CSV, convert the final JSON back to the original format and compare
if [ "$EXTENSION" = "yaml" ] || [ "$EXTENSION" = "yml" ] || [ "$EXTENSION" = "csv" ]; then
    echo "Converting final JSON back to $EXTENSION..."
    convert_file "${JSON_FILE}.2" "${INPUT_FILE}.2"
    echo "Comparing $EXTENSION files..."
    if ! compare_files "$INPUT_FILE" "${INPUT_FILE}.2"; then
        echo "$EXTENSION round-trip failed"
        exit 1
    fi
fi

echo "Round-trip conversion successful!"

# Clean up temporary files
rm -f "$JSON_FILE" "$YAML_FILE" "${JSON_FILE}.2" "${INPUT_FILE}.2"