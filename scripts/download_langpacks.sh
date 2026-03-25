#!/bin/bash
# Download all Telegram iOS language packs and merge custom strings.
# Usage: cd teamgram-server && bash scripts/download_langpacks.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$ROOT_DIR/data/langpack/ios"
CUSTOM_STRINGS="$ROOT_DIR/app/bff/langpack/internal/dao/custom_strings.json"

LANGS=(
  zh-hans zh-hant en ar be ca hr cs nl fi fr de he hu id it kk ko ms nb fa pl pt-br ro ru sr sk es sv tr uk uz vi
)

mkdir -p "$OUTPUT_DIR"

echo "=== Downloading ${#LANGS[@]} language packs ==="

for lang in "${LANGS[@]}"; do
  url="https://translations.telegram.org/${lang}/ios/export"
  outfile="$OUTPUT_DIR/${lang}.strings"
  echo -n "  $lang ... "

  http_code=$(curl -s -o "$outfile" -w "%{http_code}" "$url" --max-time 60 2>/dev/null || echo "000")

  if [ "$http_code" = "200" ]; then
    size=$(wc -c < "$outfile" | tr -d ' ')
    echo "OK (${size} bytes)"
  else
    echo "FAILED (HTTP $http_code)"
    rm -f "$outfile"
  fi
done

# Merge custom strings into each downloaded .strings file
echo ""
echo "=== Merging custom strings ==="

if [ ! -f "$CUSTOM_STRINGS" ]; then
  echo "WARNING: custom_strings.json not found at $CUSTOM_STRINGS, skipping merge"
else
  # Use python3 to parse JSON and merge
  python3 - "$CUSTOM_STRINGS" "$OUTPUT_DIR" <<'PYEOF'
import json, sys, os, re

custom_file = sys.argv[1]
output_dir = sys.argv[2]

with open(custom_file, 'r', encoding='utf-8') as f:
    custom = json.load(f)

def escape_value(v):
    """Escape a string value for Apple .strings format."""
    v = v.replace('\\', '\\\\')
    v = v.replace('"', '\\"')
    v = v.replace('\n', '\\n')
    return v

def get_existing_keys(filepath):
    """Parse a .strings file and return a set of existing keys."""
    keys = set()
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if line.startswith('"'):
                    m = re.match(r'^"([^"]+)"\s*=', line)
                    if m:
                        keys.add(m.group(1))
    except FileNotFoundError:
        pass
    return keys

def find_custom_strings(lang_code, custom):
    """Find custom strings for a language code with fallback."""
    code = lang_code.lower()
    if code in custom:
        return custom[code]
    # Try base prefix: pt-br -> pt
    if '-' in code:
        base = code.split('-')[0]
        if base in custom:
            return custom[base]
    # Fallback to English
    return custom.get('en', {})

for filename in sorted(os.listdir(output_dir)):
    if not filename.endswith('.strings'):
        continue
    lang_code = filename[:-len('.strings')]
    filepath = os.path.join(output_dir, filename)

    existing_keys = get_existing_keys(filepath)
    custom_entries = find_custom_strings(lang_code, custom)

    new_entries = {k: v for k, v in custom_entries.items() if k not in existing_keys}
    if not new_entries:
        print(f"  {lang_code}: no new custom strings to add")
        continue

    with open(filepath, 'a', encoding='utf-8') as f:
        f.write('\n\n// === Custom Strings ===\n')
        for key in sorted(new_entries.keys()):
            val = escape_value(new_entries[key])
            f.write(f'"{key}" = "{val}";\n')

    print(f"  {lang_code}: added {len(new_entries)} custom strings")

PYEOF
fi

# Create Thai language pack from English
echo ""
echo "=== Creating Thai language pack ==="
EN_FILE="$OUTPUT_DIR/en.strings"
TH_FILE="$OUTPUT_DIR/th.strings"
if [ -f "$EN_FILE" ]; then
  {
    echo "// Thai - copied from English, pending translation"
    echo ""
    cat "$EN_FILE"
  } > "$TH_FILE"

  # Replace English custom strings with Thai ones (from custom_strings.json)
  python3 - "$CUSTOM_STRINGS" "$TH_FILE" <<'PYEOF'
import json, sys

custom_file = sys.argv[1]
th_file = sys.argv[2]

with open(custom_file, 'r', encoding='utf-8') as f:
    custom = json.load(f)

# Thai custom strings = English (since no Thai translation exists yet)
th_strings = custom.get('th', custom.get('en', {}))

# Read file, no changes needed since Thai copies English custom strings
# The custom strings were already appended from English merge step
print("  th: created from English base")
PYEOF
else
  echo "  ERROR: English .strings file not found, cannot create Thai"
fi

echo ""
echo "=== Done ==="
echo "Files in $OUTPUT_DIR:"
ls -1 "$OUTPUT_DIR"/*.strings 2>/dev/null | wc -l | tr -d ' '
echo " language pack files"
