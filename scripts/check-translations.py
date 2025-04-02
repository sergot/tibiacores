#!/usr/bin/env python3

import json
import os
import sys
from typing import Dict, Set


def extract_keys(obj: dict, prefix: str = "") -> Set[str]:
    """Extract all translation keys from a nested dictionary."""
    keys = set()
    for key, value in obj.items():
        full_key = f"{prefix}.{key}" if prefix else key
        if isinstance(value, dict):
            keys.update(extract_keys(value, full_key))
        else:
            keys.add(full_key)
    return keys


def check_key_usage(key: str, frontend_dir: str) -> bool:
    """Check if a translation key is used in the frontend code."""
    extensions = [".ts", ".tsx", ".js", ".jsx", ".vue"]
    patterns = [
        f't("{key}")',
        f"t('{key}')",
        f"t(`{key}`)",
        f'useTranslation("{key}")',
        f"useTranslation('{key}')",
        f"useTranslation(`{key}`)",
        f'{{t("{key}")}}',
        f"{{t('{key}')}}",
        f"{{t(`{key}`)}}",
        f'$t("{key}")',
        f"$t('{key}')",
        f"$t(`{key}`)",
        f'{{ $t("{key}") }}',
        f"{{ $t('{key}') }}",
        f"{{ $t(`{key}`) }}",
        f'i18n.t("{key}")',
        f"i18n.t('{key}')",
        f"i18n.t(`{key}`)",
        f'i18n.global.t("{key}")',
        f"i18n.global.t('{key}')",
        f"i18n.global.t(`{key}`)",
    ]

    try:
        for root, _, files in os.walk(frontend_dir):
            for file in files:
                if any(file.endswith(ext) for ext in extensions):
                    file_path = os.path.join(root, file)
                    try:
                        with open(file_path, "r", encoding="utf-8") as f:
                            content = f.read()
                            if any(pattern in content for pattern in patterns):
                                return True
                    except Exception:
                        continue
    except Exception as e:
        print(f"Error scanning frontend directory: {e}", file=sys.stderr)
        return False
    return False


def get_value_from_path(obj: dict, path: str):
    """Get a value from a nested dictionary using a dot-separated path."""
    parts = path.split(".")
    current = obj
    for part in parts:
        if part in current:
            current = current[part]
        else:
            return None
    return current


def set_value_in_path(obj: dict, path: str, value):
    """Set a value in a nested dictionary using a dot-separated path."""
    parts = path.split(".")
    current = obj
    for i, part in enumerate(parts[:-1]):
        if part not in current:
            current[part] = {}
        current = current[part]
    current[parts[-1]] = value


def clean_translations(translations: dict, used_keys: Set[str]) -> dict:
    """Create a new translations dict with only used keys."""
    cleaned = {}
    for key in sorted(used_keys):
        value = get_value_from_path(translations, key)
        if value is not None:
            set_value_in_path(cleaned, key, value)
    return cleaned


def main():
    # Get the project root directory (parent of scripts directory)
    script_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(script_dir)
    frontend_dir = os.path.join(project_root, "frontend", "src")
    locales_dir = os.path.join(frontend_dir, "i18n", "locales")

    # Check if directories exist
    if not os.path.exists(frontend_dir):
        print(f"Error: Frontend directory not found at {frontend_dir}", file=sys.stderr)
        sys.exit(1)
    if not os.path.exists(locales_dir):
        print(f"Error: Locales directory not found at {locales_dir}", file=sys.stderr)
        sys.exit(1)

    # Initialize language dictionaries
    languages = ['en', 'de', 'es', 'pl']
    translations: Dict[str, dict] = {}
    
    # Load translation files
    for lang in languages:
        file_path = os.path.join(locales_dir, f"{lang}.json")
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                translations[lang] = json.load(f)
        except Exception as e:
            print(f"Error loading {lang}.json: {e}", file=sys.stderr)
            translations[lang] = {}
    
    # Get all keys from English file
    en_keys = extract_keys(translations['en'])
    
    # Check for unused keys in English file
    used_keys = set()
    unused_keys = set()
    for key in en_keys:
        if check_key_usage(key, frontend_dir):
            used_keys.add(key)
        else:
            unused_keys.add(key)
    
    # Print summary
    print("\nTranslation Key Summary:")
    print("=" * 50)
    print(f"Total keys in English file: {len(en_keys)}")
    print(f"Unused keys: {len(unused_keys)}")
    print(f"Used keys: {len(used_keys)}")
    print(f"Usage rate: {len(used_keys) / len(en_keys) * 100:.1f}%")
    
    # Create cleaned translation files
    for lang in languages:
        original_keys = extract_keys(translations[lang])
        cleaned = clean_translations(translations[lang], used_keys)
        cleaned_keys = extract_keys(cleaned)
        
        # Only write and report if keys were actually removed
        if len(original_keys) > len(cleaned_keys):
            output_path = os.path.join(locales_dir, f"{lang}.json")
            try:
                with open(output_path, 'w', encoding='utf-8') as f:
                    json.dump(cleaned, f, indent=2, ensure_ascii=False)
                    f.write('\n')  # Add newline at end of file
                print(f"\nCleaned {output_path} (removed {len(original_keys) - len(cleaned_keys)} unused keys)")
            except Exception as e:
                print(f"Error writing {lang}.json: {e}", file=sys.stderr)
        
    if unused_keys:
        print("\nRemoved Keys:")
        print("=" * 50)
        for key in sorted(unused_keys):
            print(f"- {key}")


if __name__ == "__main__":
    main()
