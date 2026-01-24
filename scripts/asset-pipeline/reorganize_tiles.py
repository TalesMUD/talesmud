#!/usr/bin/env python3
"""
Reorganize extracted tiles into proper category folders.
Based on visual inspection of DarkIcons0.4.png (320x640, 10x20 grid)
CORRECTED VERSION based on actual tile content
"""

import shutil
from pathlib import Path

SOURCE_DIR = Path("public/assets/img/tiles/uncategorized")
TARGET_DIR = Path("public/assets/img")

# Corrected mapping based on actual visual inspection
TILE_MAP = {
    # Row 0: Swords
    (0, 0): ("items/weapons/swords", "sword-rusty"),
    (0, 1): ("items/weapons/swords", "sword-iron"),
    (0, 2): ("items/weapons/swords", "sword-steel"),
    (0, 3): ("items/weapons/swords", "sword-sapphire"),
    (0, 4): ("items/weapons/swords", "sword-ornate"),
    (0, 5): ("items/weapons/swords", "scimitar"),
    (0, 6): ("items/weapons/swords", "broadsword"),
    (0, 7): ("items/weapons/swords", "shortsword"),
    (0, 8): ("items/weapons/swords", "sword-studded"),
    (0, 9): ("items/weapons/swords", "sword-jade"),

    # Row 1: Shields
    (1, 0): ("items/armor/shields", "shield-wooden-tall"),
    (1, 1): ("items/armor/shields", "shield-wooden-round"),
    (1, 2): ("items/armor/shields", "shield-iron-round"),
    (1, 3): ("items/armor/shields", "shield-kite-blue"),
    (1, 4): ("items/armor/shields", "shield-kite-red"),
    (1, 5): ("items/armor/shields", "shield-buckler"),
    (1, 6): ("items/armor/shields", "shield-tower"),
    (1, 7): ("items/armor/shields", "shield-ornate"),
    (1, 8): ("items/armor/shields", "shield-spiked"),
    (1, 9): ("items/armor/shields", "shield-nature"),

    # Row 2: More swords
    (2, 0): ("items/weapons/swords", "longsword-iron"),
    (2, 1): ("items/weapons/swords", "longsword-steel"),
    (2, 2): ("items/weapons/swords", "longsword-silver"),
    (2, 3): ("items/weapons/swords", "longsword-gold"),
    (2, 4): ("items/weapons/swords", "longsword-ruby"),
    (2, 5): ("items/weapons/swords", "rapier"),
    (2, 6): ("items/weapons/swords", "katana"),
    (2, 7): ("items/weapons/swords", "claymore"),
    (2, 8): ("items/weapons/swords", "flamberge"),
    (2, 9): ("items/weapons/swords", "sword-crystal"),

    # Row 3: Potions (small) and daggers
    (3, 0): ("items/consumables/potions", "potion-red-small"),
    (3, 1): ("items/consumables/potions", "potion-blue-small"),
    (3, 2): ("items/consumables/potions", "potion-green-small"),
    (3, 3): ("items/consumables/potions", "potion-yellow-small"),
    (3, 4): ("items/consumables/potions", "potion-purple-small"),
    (3, 5): ("items/weapons/daggers", "dagger-iron"),
    (3, 6): ("items/weapons/daggers", "dagger-steel"),
    (3, 7): ("items/weapons/daggers", "dagger-gold"),
    (3, 8): ("items/weapons/daggers", "dagger-ornate"),
    (3, 9): ("items/weapons/daggers", "knife"),

    # Row 4: Polearms
    (4, 0): ("items/weapons/polearms", "staff-wooden"),
    (4, 1): ("items/weapons/polearms", "staff-iron"),
    (4, 2): ("items/weapons/polearms", "spear-wooden"),
    (4, 3): ("items/weapons/polearms", "spear-iron"),
    (4, 4): ("items/weapons/polearms", "spear-steel"),
    (4, 5): ("items/weapons/polearms", "lance"),
    (4, 6): ("items/weapons/polearms", "halberd"),
    (4, 7): ("items/weapons/polearms", "pike"),
    (4, 8): ("items/weapons/polearms", "trident"),
    (4, 9): ("items/weapons/polearms", "glaive"),

    # Row 5: Axes and maces
    (5, 0): ("items/weapons/axes", "axe-wooden"),
    (5, 1): ("items/weapons/axes", "axe-iron"),
    (5, 2): ("items/weapons/axes", "axe-steel"),
    (5, 3): ("items/weapons/axes", "axe-double"),
    (5, 4): ("items/weapons/axes", "battleaxe"),
    (5, 5): ("items/weapons/maces", "mace-iron"),
    (5, 6): ("items/weapons/maces", "mace-spiked"),
    (5, 7): ("items/weapons/maces", "morningstar"),
    (5, 8): ("items/weapons/maces", "flail"),
    (5, 9): ("items/weapons/maces", "warhammer"),

    # Row 6: Bows
    (6, 0): ("items/weapons/bows", "bow-short"),
    (6, 1): ("items/weapons/bows", "bow-long"),
    (6, 2): ("items/weapons/bows", "bow-composite"),
    (6, 3): ("items/weapons/bows", "bow-recurve"),
    (6, 4): ("items/weapons/bows", "bow-elven"),
    (6, 5): ("items/weapons/bows", "bow-dark"),
    (6, 6): ("items/weapons/bows", "crossbow-light"),
    (6, 7): ("items/weapons/bows", "crossbow-heavy"),
    (6, 8): ("items/weapons/ammo", "arrows"),
    (6, 9): ("items/weapons/ammo", "bolts"),

    # Row 7: Staves and wands
    (7, 0): ("items/weapons/staves", "staff-walking"),
    (7, 1): ("items/weapons/staves", "staff-gnarled"),
    (7, 2): ("items/weapons/staves", "staff-crystal"),
    (7, 3): ("items/weapons/staves", "staff-fire"),
    (7, 4): ("items/weapons/staves", "staff-ice"),
    (7, 5): ("items/weapons/staves", "staff-nature"),
    (7, 6): ("items/weapons/wands", "wand-wooden"),
    (7, 7): ("items/weapons/wands", "wand-bone"),
    (7, 8): ("items/weapons/wands", "wand-crystal"),
    (7, 9): ("items/weapons/wands", "wand-gold"),

    # Row 8: Orbs
    (8, 0): ("items/magical/orbs", "orb-dark"),
    (8, 1): ("items/magical/orbs", "orb-red"),
    (8, 2): ("items/magical/orbs", "orb-orange"),
    (8, 3): ("items/magical/orbs", "orb-yellow"),
    (8, 4): ("items/magical/orbs", "orb-green"),
    (8, 5): ("items/magical/orbs", "orb-cyan"),
    (8, 6): ("items/magical/orbs", "orb-blue"),
    (8, 7): ("items/magical/orbs", "orb-purple"),
    (8, 8): ("items/magical/orbs", "orb-pink"),
    (8, 9): ("items/magical/orbs", "orb-white"),

    # Row 9: Books
    (9, 0): ("items/magical/books", "book-brown"),
    (9, 1): ("items/magical/books", "book-red"),
    (9, 2): ("items/magical/books", "book-blue"),
    (9, 3): ("items/magical/books", "book-green"),
    (9, 4): ("items/magical/books", "book-purple"),
    (9, 5): ("items/magical/books", "tome-ancient"),
    (9, 6): ("items/magical/books", "grimoire"),
    (9, 7): ("items/magical/books", "spellbook"),
    (9, 8): ("items/magical/books", "scroll"),
    (9, 9): ("items/magical/books", "scroll-sealed"),

    # Row 10: Potions (larger)
    (10, 0): ("items/consumables/potions", "potion-health"),
    (10, 1): ("items/consumables/potions", "potion-mana"),
    (10, 2): ("items/consumables/potions", "potion-stamina"),
    (10, 3): ("items/consumables/potions", "potion-poison"),
    (10, 4): ("items/consumables/potions", "potion-antidote"),
    (10, 5): ("items/consumables/potions", "potion-strength"),
    (10, 6): ("items/consumables/potions", "potion-speed"),
    (10, 7): ("items/consumables/potions", "potion-invisibility"),
    (10, 8): ("items/consumables/potions", "elixir"),
    (10, 9): ("items/consumables/potions", "flask-empty"),

    # Row 11: Accessories/jewelry
    (11, 0): ("items/accessories", "ring-gold"),
    (11, 1): ("items/accessories", "ring-silver"),
    (11, 2): ("items/accessories", "ring-ruby"),
    (11, 3): ("items/accessories", "amulet-gold"),
    (11, 4): ("items/accessories", "amulet-silver"),
    (11, 5): ("items/accessories", "necklace"),
    (11, 6): ("items/accessories", "pendant"),
    (11, 7): ("items/accessories", "bracelet"),
    (11, 8): ("items/accessories", "ankh"),
    (11, 9): ("items/accessories", "talisman"),

    # Row 12: Flowers/herbs and bones
    (12, 0): ("items/consumables/herbs", "flower-pink"),
    (12, 1): ("items/misc", "bones-crossed"),
    (12, 2): ("items/misc", "bones-blue"),
    (12, 3): ("items/misc", "bones-red"),
    (12, 4): ("items/consumables/herbs", "herb-green"),
    (12, 5): ("items/consumables/herbs", "herb-red"),
    (12, 6): ("items/consumables/herbs", "mushroom-brown"),
    (12, 7): ("items/consumables/herbs", "mushroom-blue"),
    (12, 8): ("items/consumables/herbs", "root"),
    (12, 9): ("items/consumables/herbs", "leaf"),

    # Row 13: Food - vegetables
    (13, 0): ("items/consumables/food", "onion"),
    (13, 1): ("items/consumables/food", "bread-loaf"),
    (13, 2): ("items/consumables/food", "cheese-wedge"),
    (13, 3): ("items/consumables/food", "chili-pepper"),
    (13, 4): ("items/consumables/food", "garlic"),
    (13, 5): ("items/consumables/food", "pumpkin"),
    (13, 6): ("items/consumables/food", "seeds"),
    (13, 7): ("items/consumables/food", "potato"),
    (13, 8): ("items/consumables/food", "tomatoes"),
    (13, 9): ("items/consumables/herbs", "clover"),

    # Row 14: Resources - meat, bags, blocks
    (14, 0): ("items/consumables/food", "fish"),
    (14, 1): ("items/misc", "bag-leather"),
    (14, 2): ("items/resources", "stone-block"),
    (14, 3): ("items/resources", "iron-block"),
    (14, 4): ("items/resources", "gold-block"),
    (14, 5): ("items/resources", "copper-block"),
    (14, 6): ("items/consumables/food", "meat-raw"),
    (14, 7): ("items/consumables/food", "meat-cooked"),
    (14, 8): ("items/consumables/food", "drumstick"),
    (14, 9): ("items/consumables/food", "ham"),

    # Row 15: Resources - meat, wood, materials
    (15, 0): ("items/consumables/food", "steak"),
    (15, 1): ("items/resources", "wood-log"),
    (15, 2): ("items/resources", "stone"),
    (15, 3): ("items/resources", "ore-iron"),
    (15, 4): ("items/resources", "ore-gold"),
    (15, 5): ("items/resources", "ore-copper"),
    (15, 6): ("items/resources", "ingot-iron"),
    (15, 7): ("items/resources", "ingot-gold"),
    (15, 8): ("items/resources", "ingot-copper"),
    (15, 9): ("items/resources", "coal"),

    # Row 16: Tools - pickaxes and shovels
    (16, 0): ("items/tools", "pickaxe-wooden"),
    (16, 1): ("items/tools", "pickaxe-iron"),
    (16, 2): ("items/tools", "pickaxe-steel"),
    (16, 3): ("items/tools", "pickaxe-gold"),
    (16, 4): ("items/tools", "shovel-wooden"),
    (16, 5): ("items/tools", "shovel-iron"),
    (16, 6): ("items/tools", "shovel-steel"),
    (16, 7): ("items/tools", "shovel-gold"),
    (16, 8): ("items/tools", "hoe-wooden"),
    (16, 9): ("items/tools", "hoe-iron"),

    # Row 17: Tools - axes and hammers
    (17, 0): ("items/tools", "axe-tool-wooden"),
    (17, 1): ("items/tools", "axe-tool-iron"),
    (17, 2): ("items/tools", "hammer-wooden"),
    (17, 3): ("items/tools", "hammer-iron"),
    (17, 4): ("items/tools", "saw"),
    (17, 5): ("items/tools", "axe-tool-steel"),
    (17, 6): ("items/tools", "fishing-rod"),
    (17, 7): ("items/tools", "bucket"),
    (17, 8): ("items/tools", "scissors"),
    (17, 9): ("items/tools", "wrench"),

    # Row 18: Misc tools and items
    (18, 0): ("items/tools", "torch"),
    (18, 1): ("items/tools", "lantern"),
    (18, 2): ("items/tools", "rope"),
    (18, 3): ("items/tools", "chain"),
    (18, 4): ("items/misc", "key-bronze"),
    (18, 5): ("items/misc", "key-silver"),
    (18, 6): ("items/misc", "key-gold"),
    (18, 7): ("items/misc", "lockpick"),
    (18, 8): ("items/misc", "compass"),
    (18, 9): ("items/misc", "map"),

    # Row 19: Empty row (all transparent)
}


def reorganize_tiles(dry_run=False):
    """Move tiles from uncategorized to proper category folders."""
    moved = 0
    skipped = 0
    empty = 0

    for (row, col), (category, name) in TILE_MAP.items():
        source_name = f"darkicons-tile-r{row:02d}-c{col:02d}.png"
        source_path = SOURCE_DIR / source_name

        if not source_path.exists():
            print(f"  SKIP: {source_name} not found")
            skipped += 1
            continue

        # Check if tile is empty (very small file = likely transparent)
        if source_path.stat().st_size < 200:
            print(f"  EMPTY: {source_name} (likely transparent)")
            empty += 1
            continue

        target_dir = TARGET_DIR / category
        target_path = target_dir / f"{name}.png"

        if dry_run:
            print(f"  {source_name} -> {target_path}")
        else:
            target_dir.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source_path, target_path)
            print(f"  {source_name} -> {target_path}")

        moved += 1

    print()
    print(f"Moved: {moved}")
    print(f"Skipped: {skipped}")
    print(f"Empty: {empty}")


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Reorganize tiles into categories")
    parser.add_argument("--dry-run", action="store_true", help="Just show what would be done")
    args = parser.parse_args()

    print("Reorganizing tiles...")
    print(f"Source: {SOURCE_DIR}")
    print(f"Target: {TARGET_DIR}")
    print()

    reorganize_tiles(dry_run=args.dry_run)


if __name__ == "__main__":
    main()
