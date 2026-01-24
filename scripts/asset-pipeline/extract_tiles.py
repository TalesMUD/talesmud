#!/usr/bin/env python3
"""
Tile Extraction Script for TalesMUD
Extracts 32x32 tiles from sprite sheets and categorizes them into folders.
"""

import os
import sys
from pathlib import Path
from PIL import Image

# Base output directory
OUTPUT_BASE = Path("public/assets/img")

# Tile categorization map for DarkIcons0.4.png (320x640 = 10 cols x 20 rows)
# Format: (row, col): ("category/subcategory", "name")
DARKICONS_MAP = {
    # Row 0: Shields
    (0, 0): ("items/armor/shields", "wooden-shield"),
    (0, 1): ("items/armor/shields", "round-shield"),
    (0, 2): ("items/armor/shields", "kite-shield-blue"),
    (0, 3): ("items/armor/shields", "ornate-shield"),
    (0, 4): ("items/armor/shields", "buckler"),
    (0, 5): ("items/armor/shields", "tower-shield"),
    (0, 6): ("items/armor/shields", "spiked-shield"),
    (0, 7): ("items/armor/shields", "royal-shield"),
    (0, 8): ("items/armor/shields", "elven-shield"),
    (0, 9): ("items/armor/shields", "nature-shield"),

    # Row 1: More shields and accessories
    (1, 0): ("items/armor/shields", "iron-shield"),
    (1, 1): ("items/armor/shields", "steel-shield"),
    (1, 2): ("items/armor/shields", "gem-shield"),
    (1, 3): ("items/armor/shields", "knight-shield"),
    (1, 4): ("items/armor/shields", "dark-shield"),
    (1, 5): ("items/armor/shields", "skull-shield"),
    (1, 6): ("items/accessories", "pendant"),
    (1, 7): ("items/accessories", "amulet"),
    (1, 8): ("items/accessories", "ring-gold"),
    (1, 9): ("items/accessories", "ring-silver"),

    # Row 2: Swords
    (2, 0): ("items/weapons/swords", "short-sword"),
    (2, 1): ("items/weapons/swords", "long-sword"),
    (2, 2): ("items/weapons/swords", "broad-sword"),
    (2, 3): ("items/weapons/swords", "bastard-sword"),
    (2, 4): ("items/weapons/swords", "claymore"),
    (2, 5): ("items/weapons/swords", "rapier"),
    (2, 6): ("items/weapons/swords", "scimitar"),
    (2, 7): ("items/weapons/swords", "katana"),
    (2, 8): ("items/weapons/swords", "flamberge"),
    (2, 9): ("items/weapons/swords", "crystal-sword"),

    # Row 3: More swords and daggers
    (3, 0): ("items/weapons/swords", "iron-sword"),
    (3, 1): ("items/weapons/swords", "steel-sword"),
    (3, 2): ("items/weapons/swords", "silver-sword"),
    (3, 3): ("items/weapons/swords", "golden-sword"),
    (3, 4): ("items/weapons/swords", "enchanted-sword"),
    (3, 5): ("items/weapons/daggers", "dagger"),
    (3, 6): ("items/weapons/daggers", "knife"),
    (3, 7): ("items/weapons/daggers", "stiletto"),
    (3, 8): ("items/weapons/daggers", "kris"),
    (3, 9): ("items/weapons/daggers", "throwing-knife"),

    # Row 4: Polearms
    (4, 0): ("items/weapons/polearms", "spear"),
    (4, 1): ("items/weapons/polearms", "lance"),
    (4, 2): ("items/weapons/polearms", "pike"),
    (4, 3): ("items/weapons/polearms", "halberd"),
    (4, 4): ("items/weapons/polearms", "glaive"),
    (4, 5): ("items/weapons/polearms", "trident"),
    (4, 6): ("items/weapons/polearms", "pitchfork"),
    (4, 7): ("items/weapons/polearms", "war-scythe"),
    (4, 8): ("items/weapons/polearms", "javelin"),
    (4, 9): ("items/weapons/polearms", "partisan"),

    # Row 5: More polearms and staves
    (5, 0): ("items/weapons/polearms", "iron-spear"),
    (5, 1): ("items/weapons/polearms", "steel-lance"),
    (5, 2): ("items/weapons/polearms", "golden-trident"),
    (5, 3): ("items/weapons/staves", "wooden-staff"),
    (5, 4): ("items/weapons/staves", "iron-staff"),
    (5, 5): ("items/weapons/staves", "quarter-staff"),
    (5, 6): ("items/weapons/staves", "bo-staff"),
    (5, 7): ("items/weapons/staves", "walking-stick"),
    (5, 8): ("items/weapons/staves", "ornate-staff"),
    (5, 9): ("items/weapons/staves", "battle-staff"),

    # Row 6: Bows
    (6, 0): ("items/weapons/bows", "short-bow"),
    (6, 1): ("items/weapons/bows", "long-bow"),
    (6, 2): ("items/weapons/bows", "composite-bow"),
    (6, 3): ("items/weapons/bows", "recurve-bow"),
    (6, 4): ("items/weapons/bows", "elven-bow"),
    (6, 5): ("items/weapons/bows", "hunting-bow"),
    (6, 6): ("items/weapons/bows", "war-bow"),
    (6, 7): ("items/weapons/bows", "dark-bow"),
    (6, 8): ("items/weapons/bows", "golden-bow"),
    (6, 9): ("items/weapons/bows", "enchanted-bow"),

    # Row 7: Accessories and magical items
    (7, 0): ("items/weapons/bows", "crossbow"),
    (7, 1): ("items/weapons/bows", "heavy-crossbow"),
    (7, 2): ("items/weapons/ammo", "arrow"),
    (7, 3): ("items/weapons/ammo", "bolt"),
    (7, 4): ("items/accessories", "dreamcatcher"),
    (7, 5): ("items/accessories", "ankh"),
    (7, 6): ("items/accessories", "holy-symbol"),
    (7, 7): ("items/accessories", "talisman"),
    (7, 8): ("items/accessories", "charm"),
    (7, 9): ("items/accessories", "relic"),

    # Row 8: Orbs and spheres
    (8, 0): ("items/magical/orbs", "crystal-ball"),
    (8, 1): ("items/magical/orbs", "fire-orb"),
    (8, 2): ("items/magical/orbs", "earth-orb"),
    (8, 3): ("items/magical/orbs", "water-orb"),
    (8, 4): ("items/magical/orbs", "air-orb"),
    (8, 5): ("items/magical/orbs", "dark-orb"),
    (8, 6): ("items/magical/orbs", "light-orb"),
    (8, 7): ("items/magical/orbs", "nature-orb"),
    (8, 8): ("items/magical/orbs", "arcane-orb"),
    (8, 9): ("items/magical/orbs", "soul-orb"),

    # Row 9: Books and tomes
    (9, 0): ("items/magical/books", "spellbook-red"),
    (9, 1): ("items/magical/books", "spellbook-blue"),
    (9, 2): ("items/magical/books", "spellbook-green"),
    (9, 3): ("items/magical/books", "tome-ancient"),
    (9, 4): ("items/magical/books", "tome-dark"),
    (9, 5): ("items/magical/books", "grimoire"),
    (9, 6): ("items/magical/books", "scroll-case"),
    (9, 7): ("items/magical/books", "journal"),
    (9, 8): ("items/magical/books", "codex"),
    (9, 9): ("items/magical/books", "ledger"),

    # Row 10: Potions
    (10, 0): ("items/consumables/potions", "health-potion"),
    (10, 1): ("items/consumables/potions", "mana-potion"),
    (10, 2): ("items/consumables/potions", "stamina-potion"),
    (10, 3): ("items/consumables/potions", "poison"),
    (10, 4): ("items/consumables/potions", "antidote"),
    (10, 5): ("items/consumables/potions", "elixir"),
    (10, 6): ("items/misc", "hourglass"),
    (10, 7): ("items/misc", "hourglass-gold"),
    (10, 8): ("items/misc", "chalice"),
    (10, 9): ("items/misc", "goblet"),

    # Row 11: More misc items
    (11, 0): ("items/misc", "candelabra"),
    (11, 1): ("items/misc", "torch"),
    (11, 2): ("items/misc", "lantern"),
    (11, 3): ("items/misc", "key-bronze"),
    (11, 4): ("items/misc", "key-silver"),
    (11, 5): ("items/misc", "key-gold"),
    (11, 6): ("items/misc", "lock"),
    (11, 7): ("items/misc", "chest-small"),
    (11, 8): ("items/misc", "rope"),
    (11, 9): ("items/misc", "whip"),

    # Row 12: Food - flowers and fruits
    (12, 0): ("items/consumables/food", "flower-pink"),
    (12, 1): ("items/consumables/food", "herb-green"),
    (12, 2): ("items/consumables/food", "herb-red"),
    (12, 3): ("items/consumables/food", "mushroom"),
    (12, 4): ("items/consumables/food", "apple-red"),
    (12, 5): ("items/consumables/food", "apple-green"),
    (12, 6): ("items/consumables/food", "cherry"),
    (12, 7): ("items/consumables/food", "grapes"),
    (12, 8): ("items/consumables/food", "bread"),
    (12, 9): ("items/consumables/food", "bread-loaf"),

    # Row 13: More food
    (13, 0): ("items/consumables/food", "cheese-wheel"),
    (13, 1): ("items/consumables/food", "cheese-wedge"),
    (13, 2): ("items/consumables/food", "meat-raw"),
    (13, 3): ("items/consumables/food", "meat-cooked"),
    (13, 4): ("items/consumables/food", "fish"),
    (13, 5): ("items/consumables/food", "onion"),
    (13, 6): ("items/consumables/food", "carrot"),
    (13, 7): ("items/consumables/food", "pumpkin"),
    (13, 8): ("items/consumables/food", "egg"),
    (13, 9): ("items/consumables/food", "egg-basket"),

    # Row 14: Resources and materials
    (14, 0): ("items/resources", "acorn"),
    (14, 1): ("items/resources", "seed-bag"),
    (14, 2): ("items/resources", "hay-bale"),
    (14, 3): ("items/resources", "wood-log"),
    (14, 4): ("items/resources", "stone-block"),
    (14, 5): ("items/resources", "gold-ingot"),
    (14, 6): ("items/resources", "iron-ingot"),
    (14, 7): ("items/resources", "copper-ingot"),
    (14, 8): ("items/resources", "silver-ingot"),
    (14, 9): ("items/resources", "gem-ruby"),

    # Row 15: More resources
    (15, 0): ("items/resources", "gem-emerald"),
    (15, 1): ("items/resources", "gem-sapphire"),
    (15, 2): ("items/resources", "gem-diamond"),
    (15, 3): ("items/resources", "coal"),
    (15, 4): ("items/resources", "ore-iron"),
    (15, 5): ("items/resources", "ore-gold"),
    (15, 6): ("items/resources", "leather"),
    (15, 7): ("items/resources", "cloth"),
    (15, 8): ("items/resources", "feather"),
    (15, 9): ("items/resources", "bone"),

    # Row 16: Tools
    (16, 0): ("items/tools", "shovel-wood"),
    (16, 1): ("items/tools", "shovel-iron"),
    (16, 2): ("items/tools", "shovel-gold"),
    (16, 3): ("items/tools", "pickaxe-wood"),
    (16, 4): ("items/tools", "pickaxe-iron"),
    (16, 5): ("items/tools", "pickaxe-gold"),
    (16, 6): ("items/tools", "axe-wood"),
    (16, 7): ("items/tools", "axe-iron"),
    (16, 8): ("items/tools", "axe-gold"),
    (16, 9): ("items/tools", "hammer"),

    # Row 17: More tools
    (17, 0): ("items/tools", "hammer-war"),
    (17, 1): ("items/tools", "mallet"),
    (17, 2): ("items/tools", "saw"),
    (17, 3): ("items/tools", "chisel"),
    (17, 4): ("items/tools", "tongs"),
    (17, 5): ("items/tools", "anvil"),
    (17, 6): ("items/tools", "bucket"),
    (17, 7): ("items/tools", "watering-can"),
    (17, 8): ("items/tools", "fishing-rod"),
    (17, 9): ("items/tools", "net"),

    # Row 18: Magical staves
    (18, 0): ("items/weapons/staves", "magic-staff-fire"),
    (18, 1): ("items/weapons/staves", "magic-staff-ice"),
    (18, 2): ("items/weapons/staves", "magic-staff-lightning"),
    (18, 3): ("items/weapons/staves", "magic-staff-nature"),
    (18, 4): ("items/weapons/staves", "magic-staff-dark"),
    (18, 5): ("items/weapons/staves", "magic-staff-light"),
    (18, 6): ("items/weapons/staves", "wizard-staff"),
    (18, 7): ("items/weapons/staves", "druid-staff"),
    (18, 8): ("items/weapons/staves", "necro-staff"),
    (18, 9): ("items/weapons/staves", "holy-staff"),

    # Row 19: Wands
    (19, 0): ("items/weapons/wands", "wand-basic"),
    (19, 1): ("items/weapons/wands", "wand-fire"),
    (19, 2): ("items/weapons/wands", "wand-ice"),
    (19, 3): ("items/weapons/wands", "wand-lightning"),
    (19, 4): ("items/weapons/wands", "wand-nature"),
    (19, 5): ("items/weapons/wands", "wand-dark"),
    (19, 6): ("items/weapons/wands", "wand-light"),
    (19, 7): ("items/weapons/wands", "wand-crystal"),
    (19, 8): ("items/weapons/wands", "wand-bone"),
    (19, 9): ("items/weapons/wands", "wand-royal"),
}


def extract_tiles(
    input_path: str,
    tile_width: int = 32,
    tile_height: int = 32,
    tile_map: dict = None,
    output_base: Path = OUTPUT_BASE,
    prefix: str = "",
    dry_run: bool = False
):
    """
    Extract tiles from a sprite sheet.

    Args:
        input_path: Path to the sprite sheet image
        tile_width: Width of each tile in pixels
        tile_height: Height of each tile in pixels
        tile_map: Dictionary mapping (row, col) to ("category", "name")
        output_base: Base output directory
        prefix: Prefix for auto-generated names if no map provided
        dry_run: If True, just print what would be done
    """
    img = Image.open(input_path)
    width, height = img.size

    cols = width // tile_width
    rows = height // tile_height

    print(f"Image: {input_path}")
    print(f"Size: {width}x{height}")
    print(f"Grid: {cols} cols x {rows} rows = {cols * rows} tiles")
    print(f"Output: {output_base}")
    print()

    extracted = 0
    skipped = 0

    for row in range(rows):
        for col in range(cols):
            # Extract tile
            left = col * tile_width
            top = row * tile_height
            right = left + tile_width
            bottom = top + tile_height

            tile = img.crop((left, top, right, bottom))

            # Check if tile is empty (fully transparent)
            if tile.mode == 'RGBA':
                alpha = tile.split()[3]
                if alpha.getextrema() == (0, 0):
                    skipped += 1
                    continue

            # Get category and name from map or generate
            if tile_map and (row, col) in tile_map:
                category, name = tile_map[(row, col)]
            else:
                category = "uncategorized"
                name = f"{prefix}tile-r{row:02d}-c{col:02d}"

            # Build output path
            output_dir = output_base / category
            output_file = output_dir / f"{name}.png"

            if dry_run:
                print(f"  [{row},{col}] -> {output_file}")
            else:
                output_dir.mkdir(parents=True, exist_ok=True)
                tile.save(output_file, "PNG")
                print(f"  [{row},{col}] -> {output_file}")

            extracted += 1

    print()
    print(f"Extracted: {extracted} tiles")
    print(f"Skipped (empty): {skipped} tiles")

    return extracted


def main():
    import argparse

    parser = argparse.ArgumentParser(description="Extract tiles from sprite sheets")
    parser.add_argument("input", help="Input sprite sheet image")
    parser.add_argument("--tile-size", type=int, default=32, help="Tile size (default: 32)")
    parser.add_argument("--output", type=str, default=str(OUTPUT_BASE), help="Output base directory")
    parser.add_argument("--prefix", type=str, default="", help="Prefix for auto-generated names")
    parser.add_argument("--dry-run", action="store_true", help="Just show what would be done")
    parser.add_argument("--no-map", action="store_true", help="Don't use built-in tile maps")

    args = parser.parse_args()

    # Select appropriate map based on input file
    tile_map = None
    if not args.no_map:
        if "DarkIcons" in args.input:
            tile_map = DARKICONS_MAP
            print("Using DarkIcons tile map")

    extract_tiles(
        args.input,
        tile_width=args.tile_size,
        tile_height=args.tile_size,
        tile_map=tile_map,
        output_base=Path(args.output),
        prefix=args.prefix,
        dry_run=args.dry_run,
    )


if __name__ == "__main__":
    main()
