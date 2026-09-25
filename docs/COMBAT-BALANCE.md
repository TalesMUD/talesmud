# Combat balance — level gap table

Player level is 10. Gap is enemy level minus player level. Trash / elite / boss use the level-scaled bodies in `CreateScaledEnemy` (easy / hard / boss). Appropriate gear is class starter plus a small per-level bump. Good gear is about 2.2× that. Hit, crit, and the level gap come from `config/combat_balance.yaml` `level_gap`. `class_balance` then scales each class's damage dealt and taken.

Targets: level-appropriate trash at gap 0 wins about 95%+, elites 75–90%, bosses 50–65%. Bosses at +3 with good gear about 50%. +5 with appropriate gear stays under 15%.

## Warrior — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 71% | 54% |
| trash | rounds | 4.8 | 5.0 | 5.8 | 7.0 | 7.0 | 7.2 | 8.6 | 9.0 | 10.1 |
| trash | HP left | 98% | 98% | 95% | 90% | 85% | 83% | 67% | 55% | 45% |
| elite | win% | 100% | 100% | 100% | 96% | 88% | 71% | 12% | 0% | 0% |
| elite | rounds | 11.0 | 12.4 | 14.3 | 17.4 | 18.1 | 19.8 | 17.5 | 15.4 | 9.3 |
| elite | HP left | 94% | 90% | 85% | 70% | 44% | 37% | 38% | — | — |
| boss | win% | 100% | 100% | 83% | 50% | 21% | 0% | 0% | 0% | 0% |
| boss | rounds | 21.2 | 25.8 | 28.8 | 28.2 | 21.6 | 21.3 | 12.7 | 10.8 | 9.9 |
| boss | HP left | 85% | 71% | 51% | 41% | 28% | — | — | — | — |

## Warrior — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% |
| trash | rounds | 4.0 | 5.0 | 5.1 | 5.3 | 5.2 | 6.7 | 7.5 | 8.5 | 9.7 |
| trash | HP left | 99% | 98% | 98% | 96% | 97% | 93% | 93% | 90% | 83% |
| elite | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 88% | 79% |
| elite | rounds | 8.7 | 11.0 | 11.6 | 13.8 | 14.5 | 16.3 | 21.4 | 25.1 | 30.0 |
| elite | HP left | 97% | 94% | 84% | 86% | 82% | 75% | 74% | 58% | 65% |
| boss | win% | 100% | 100% | 100% | 96% | 88% | 71% | 58% | 4% | 0% |
| boss | rounds | 17.2 | 19.9 | 23.2 | 28.8 | 29.9 | 32.9 | 41.1 | 55.4 | 45.2 |
| boss | HP left | 95% | 82% | 80% | 61% | 50% | 49% | 35% | 8% | — |

## Rogue — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 96% | 100% | 100% | 100% | 100% | 62% |
| trash | rounds | 4.0 | 4.3 | 4.1 | 5.0 | 3.0 | 3.1 | 3.1 | 4.4 | 3.5 |
| trash | HP left | 97% | 96% | 94% | 79% | 88% | 78% | 83% | 61% | 83% |
| elite | win% | 100% | 96% | 100% | 79% | 92% | 83% | 38% | 21% | 8% |
| elite | rounds | 7.9 | 8.7 | 9.9 | 11.5 | 5.8 | 6.9 | 5.8 | 5.5 | 5.1 |
| elite | HP left | 93% | 87% | 90% | 70% | 80% | 58% | 60% | 54% | 38% |
| boss | win% | 100% | 96% | 75% | 62% | 50% | 21% | 25% | 0% | 0% |
| boss | rounds | 14.3 | 16.0 | 16.6 | 18.5 | 9.0 | 9.2 | 8.1 | 5.4 | 5.5 |
| boss | HP left | 73% | 56% | 64% | 28% | 78% | 64% | 79% | — | — |

## Rogue — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% |
| trash | rounds | 3.0 | 3.8 | 4.0 | 3.8 | 3.0 | 3.1 | 3.1 | 3.2 | 4.8 |
| trash | HP left | 99% | 98% | 96% | 91% | 95% | 98% | 95% | 97% | 79% |
| elite | win% | 100% | 100% | 100% | 92% | 100% | 100% | 100% | 92% | 88% |
| elite | rounds | 6.5 | 7.9 | 8.0 | 8.8 | 6.0 | 5.9 | 6.8 | 8.5 | 10.8 |
| elite | HP left | 94% | 89% | 82% | 83% | 88% | 83% | 90% | 76% | 58% |
| boss | win% | 100% | 92% | 88% | 79% | 96% | 83% | 46% | 33% | 12% |
| boss | rounds | 11.8 | 13.0 | 15.3 | 16.4 | 9.3 | 11.7 | 9.9 | 11.7 | 10.5 |
| boss | HP left | 84% | 73% | 73% | 66% | 79% | 53% | 100% | 100% | 100% |

## Ranger — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 88% | 79% | 42% | 12% |
| trash | rounds | 4.0 | 4.0 | 5.0 | 5.0 | 6.0 | 6.8 | 6.9 | 6.7 | 7.0 |
| trash | HP left | 99% | 96% | 93% | 95% | 79% | 67% | 54% | 49% | 61% |
| elite | win% | 100% | 100% | 92% | 92% | 71% | 21% | 0% | 0% | 0% |
| elite | rounds | 7.8 | 9.1 | 9.8 | 13.3 | 14.2 | 10.7 | 9.1 | 9.5 | 7.7 |
| elite | HP left | 92% | 88% | 84% | 53% | 30% | 57% | — | — | — |
| boss | win% | 92% | 92% | 67% | 25% | 12% | 4% | 0% | 0% | 0% |
| boss | rounds | 14.5 | 19.2 | 20.5 | 16.7 | 15.4 | 14.5 | 10.5 | 5.4 | 5.0 |
| boss | HP left | 79% | 62% | 47% | 58% | 32% | 8% | — | — | — |

## Ranger — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 88% | 96% |
| trash | rounds | 3.0 | 3.9 | 4.0 | 5.2 | 5.2 | 5.7 | 7.2 | 7.3 | 8.8 |
| trash | HP left | 100% | 98% | 97% | 93% | 93% | 86% | 86% | 73% | 78% |
| elite | win% | 100% | 100% | 100% | 100% | 100% | 88% | 79% | 75% | 67% |
| elite | rounds | 6.5 | 8.1 | 9.1 | 10.7 | 12.7 | 14.2 | 15.4 | 19.7 | 24.8 |
| elite | HP left | 95% | 87% | 84% | 83% | 81% | 67% | 68% | 70% | 63% |
| boss | win% | 100% | 96% | 96% | 75% | 50% | 67% | 33% | 21% | 4% |
| boss | rounds | 13.2 | 14.1 | 17.6 | 18.5 | 20.1 | 27.7 | 24.0 | 31.4 | 46.9 |
| boss | HP left | 89% | 77% | 64% | 69% | 58% | 59% | 50% | 47% | 1% |

## Mage — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 88% | 92% | 75% | 12% |
| trash | rounds | 3.0 | 3.0 | 3.0 | 3.0 | 3.1 | 4.1 | 4.1 | 4.2 | 3.5 |
| trash | HP left | 91% | 86% | 86% | 75% | 72% | 65% | 48% | 32% | 39% |
| elite | win% | 100% | 100% | 100% | 92% | 79% | 71% | 8% | 4% | 0% |
| elite | rounds | 4.0 | 4.9 | 5.0 | 5.9 | 5.8 | 6.1 | 4.8 | 4.6 | 5.0 |
| elite | HP left | 86% | 74% | 74% | 44% | 48% | 32% | 27% | 23% | — |
| boss | win% | 100% | 88% | 67% | 58% | 17% | 0% | 0% | 0% | 0% |
| boss | rounds | 6.9 | 6.8 | 7.2 | 8.6 | 6.8 | 6.5 | 6.7 | 3.3 | 2.9 |
| boss | HP left | 75% | 57% | 35% | 34% | 36% | — | — | — | — |

## Mage — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 88% | 79% |
| trash | rounds | 3.0 | 3.0 | 3.0 | 3.0 | 3.2 | 4.1 | 4.2 | 4.6 | 5.7 |
| trash | HP left | 99% | 99% | 95% | 98% | 96% | 85% | 88% | 71% | 52% |
| elite | win% | 100% | 100% | 100% | 100% | 96% | 96% | 88% | 75% | 33% |
| elite | rounds | 4.0 | 4.9 | 5.0 | 6.0 | 6.0 | 6.3 | 7.3 | 8.1 | 7.8 |
| elite | HP left | 96% | 95% | 98% | 79% | 90% | 70% | 73% | 64% | 72% |
| boss | win% | 100% | 100% | 96% | 88% | 88% | 71% | 50% | 12% | 0% |
| boss | rounds | 6.8 | 6.9 | 8.0 | 8.5 | 10.1 | 11.2 | 10.4 | 9.8 | 8.1 |
| boss | HP left | 87% | 93% | 91% | 84% | 88% | 66% | 61% | 28% | — |

## How to read this sample

One 24-iteration run per cell (player level 10). Win% moves several points between runs. An 80-iteration check of the same config is the steadier read of the three cells A4 left open:

| Cell | A4 sample | 80-iteration check |
| --- | --- | --- |
| Warrior, appropriate gear, boss, gap 0 | 75% | 59% |
| Rogue, good gear, boss, gap +3 | 21% | 51% |
| Mage, appropriate gear, elite, gap 0 | 8% | 74% |
| Mage, appropriate gear, boss, gap 0 | 0% | 54% |
| Mage, good gear, boss, gap +3 | 4% | 52% |

`class_balance` in `config/combat_balance.yaml` scales damage after `level_gap` and before a crit. `behind_dealt` applies only when that class is the lower level, so an even boss and a fight three levels up can be tuned apart. The scaled boss body in `CreateScaledEnemy` is `220 + 23*level` hit points. Content bosses still use `CreateEnemy` and the duration bands.

- Warrior even-fight damage is unchanged, so at-level trash duration stays in the old windows. The thicker boss is what pulls an appropriate-gear at-level boss into the 50–65% band. `behind_dealt` 1.20 keeps a good-gear boss at +3 near 60%.
- Rogue even-fight damage is 1.35×. Uphill (`behind_dealt` 2.35) is what moves the good-gear +3 boss to about 50%. Appropriate-gear +5 stays at 0%.
- Mage damage dealt is 2.65× and damage taken is 0.46×. Cloth still ends those wins on low HP. Trash at gap 0 is a clear. Elite and boss at gap 0, and a good-gear boss at +3, land in the same bands as the melee classes.
- Ranger damage dealt is 1.26× so an at-level boss is no longer a one-sided loss. Content boss fights still last at least 12 rounds.
- Appropriate-gear elites and bosses at +5 stay under 15%.

`level_gap` per level of attacker advantage is unchanged: hit +3.5%, crit +1%, damage dealt +3.5%, damage taken +2%, clamped at ±6. Gap 0 does not change the level-gap term.
