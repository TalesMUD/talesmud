#!/usr/bin/env python3
"""Crop 1:1 head-and-shoulders bust portraits from full-body NPC/enemy sprites.

Output: {templateId}-bust.png next to source sprites or in portraits/ dir.
Skips files that already exist (use AI-generated busts for key NPCs).
"""

from __future__ import annotations

import argparse
import os
from pathlib import Path

from PIL import Image


def content_bbox(im: Image.Image) -> tuple[int, int, int, int] | None:
    if im.mode != "RGBA":
        im = im.convert("RGBA")
    return im.getbbox()


def crop_bust(im: Image.Image, out_size: int = 512) -> Image.Image:
    im = im.convert("RGBA")
    bbox = content_bbox(im)
    if not bbox:
        return im.resize((out_size, out_size), Image.Resampling.LANCZOS)

    left, top, right, bottom = bbox
    cw = right - left
    ch = bottom - top
    cx = (left + right) / 2

    # Head-and-shoulders: upper ~42% of figure, square crop centered on torso
    bust_h = max(cw, int(ch * 0.42))
    bust_w = bust_h
    bust_top = top
    bust_left = int(cx - bust_w / 2)
    bust_right = bust_left + bust_w
    bust_bottom = bust_top + bust_h

    # Clamp to image bounds, shift if needed
    if bust_left < 0:
        bust_right -= bust_left
        bust_left = 0
    if bust_top < 0:
        bust_bottom -= bust_top
        bust_top = 0
    if bust_right > im.width:
        shift = bust_right - im.width
        bust_left = max(0, bust_left - shift)
        bust_right = im.width
    if bust_bottom > im.height:
        bust_bottom = im.height

    crop = im.crop((bust_left, bust_top, bust_right, bust_bottom))
    return crop.resize((out_size, out_size), Image.Resampling.LANCZOS)


def process_dir(src_dir: Path, out_dir: Path, skip_existing: bool = True) -> int:
    count = 0
    out_dir.mkdir(parents=True, exist_ok=True)
    for path in sorted(src_dir.glob("*.png")):
        if path.name.endswith("-bust.png"):
            continue
        template_id = path.stem
        out_path = out_dir / f"{template_id}-bust.png"
        if skip_existing and out_path.exists():
            continue
        im = Image.open(path)
        bust = crop_bust(im)
        bust.save(out_path, "PNG")
        count += 1
        print(f"  {template_id} -> {out_path.name}")
    return count


def main() -> None:
    parser = argparse.ArgumentParser(description="Generate bust portraits from full-body sprites")
    parser.add_argument(
        "--import-root",
        default="import/mvp-rpg-1/assets/images/sprites",
        help="Sprites root under repo",
    )
    parser.add_argument("--uploads", default="uploads/portraits", help="Runtime portraits dir")
    parser.add_argument("--force", action="store_true", help="Overwrite existing busts")
    args = parser.parse_args()

    root = Path(args.import_root)
    out = root / "portraits"
    uploads = Path(args.uploads)

    total = 0
    for sub in ("npcs", "enemies"):
        src = root / sub
        if src.is_dir():
            print(f"Processing {src}...")
            total += process_dir(src, out, skip_existing=not args.force)

    print(f"Generated {total} bust(s) in {out}")

    uploads.mkdir(parents=True, exist_ok=True)
    copied = 0
    for bust in out.glob("*-bust.png"):
        dest = uploads / bust.name
        if not dest.exists() or args.force:
            dest.write_bytes(bust.read_bytes())
            copied += 1
    print(f"Copied {copied} bust(s) to {uploads}")


if __name__ == "__main__":
    main()
