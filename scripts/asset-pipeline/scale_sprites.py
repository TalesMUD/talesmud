#!/usr/bin/env python3
"""
Scale 16x16 sprites to 32x32 using nearest-neighbor interpolation.
Preserves pixel art look.
"""

import sys
from pathlib import Path
from PIL import Image

def scale_sprites(input_dir: Path, scale_factor: int = 2, dry_run: bool = False):
    """Scale all PNG files in directory (recursively) by scale_factor."""

    png_files = list(input_dir.rglob("*.png"))
    print(f"Found {len(png_files)} PNG files in {input_dir}")

    scaled = 0
    skipped = 0
    errors = 0

    for png_path in png_files:
        try:
            img = Image.open(png_path)
            width, height = img.size

            # Skip if already at target size or larger
            target_width = width * scale_factor
            target_height = height * scale_factor

            if width >= 32 and height >= 32:
                print(f"  SKIP: {png_path} already {width}x{height}")
                skipped += 1
                continue

            if dry_run:
                print(f"  {png_path}: {width}x{height} -> {target_width}x{target_height}")
            else:
                # Use NEAREST for pixel art (no interpolation)
                scaled_img = img.resize((target_width, target_height), Image.NEAREST)
                scaled_img.save(png_path, "PNG")
                print(f"  {png_path}: {width}x{height} -> {target_width}x{target_height}")

            scaled += 1

        except Exception as e:
            print(f"  ERROR: {png_path}: {e}")
            errors += 1

    print()
    print(f"Scaled: {scaled}")
    print(f"Skipped: {skipped}")
    print(f"Errors: {errors}")


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Scale pixel art sprites")
    parser.add_argument("input_dir", help="Directory containing sprites")
    parser.add_argument("--scale", type=int, default=2, help="Scale factor (default: 2)")
    parser.add_argument("--dry-run", action="store_true", help="Just show what would be done")

    args = parser.parse_args()

    input_path = Path(args.input_dir)
    if not input_path.exists():
        print(f"Error: {input_path} does not exist")
        sys.exit(1)

    scale_sprites(input_path, args.scale, args.dry_run)


if __name__ == "__main__":
    main()
