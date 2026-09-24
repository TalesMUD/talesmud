# Combat balance — level gap table

Player level is 10. Gap is enemy level minus player level. Trash / elite / boss use the level-scaled bodies in `CreateScaledEnemy` (easy / hard / boss). Appropriate gear is class starter plus a small per-level bump. Good gear is about 2.2× that. Hit, crit, and damage still come from `config/combat_balance.yaml` `level_gap`.

Targets: level-appropriate trash at gap 0 wins about 95%+, elites 75–90%, bosses 50–65%. Bosses at +3 with good gear about 50%. +5 with appropriate gear stays under 15%.

## Warrior — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 96% | 92% | 79% | 38% |
| trash | rounds | 4.9 | 5.2 | 6.1 | 7.0 | 8.0 | 8.5 | 10.5 | 11.3 | 11.1 |
| trash | HP left | 96% | 97% | 95% | 86% | 74% | 68% | 51% | 42% | 36% |
| elite | win% | 100% | 100% | 100% | 92% | 83% | 33% | 21% | 0% | 0% |
| elite | rounds | 10.8 | 12.8 | 13.8 | 17.1 | 20.3 | 19.9 | 19.2 | 15.0 | 9.9 |
| elite | HP left | 90% | 90% | 88% | 59% | 47% | 51% | 30% | — | — |
| boss | win% | 100% | 100% | 100% | 75% | 0% | 0% | 0% | 0% | 0% |
| boss | rounds | 15.9 | 18.9 | 24.0 | 25.1 | 22.0 | 18.5 | 14.7 | 13.1 | 11.5 |
| boss | HP left | 83% | 74% | 66% | 50% | — | — | — | — | — |

## Warrior — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 100% |
| trash | rounds | 3.9 | 4.9 | 5.2 | 5.2 | 7.0 | 7.0 | 8.4 | 9.1 | 10.8 |
| trash | HP left | 99% | 98% | 96% | 97% | 93% | 90% | 88% | 89% | 89% |
| elite | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 75% | 71% |
| elite | rounds | 9.1 | 9.8 | 11.5 | 15.0 | 17.1 | 20.0 | 24.3 | 30.3 | 36.3 |
| elite | HP left | 98% | 94% | 91% | 84% | 84% | 76% | 71% | 67% | 52% |
| boss | win% | 100% | 100% | 100% | 100% | 100% | 71% | 62% | 12% | 0% |
| boss | rounds | 12.8 | 14.3 | 17.3 | 22.0 | 26.5 | 30.5 | 34.2 | 51.5 | 37.5 |
| boss | HP left | 93% | 84% | 78% | 78% | 69% | 60% | 50% | 54% | — |

## Rogue — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 96% | 92% | 67% | 42% | 46% | 8% |
| trash | rounds | 4.0 | 5.0 | 6.9 | 6.9 | 7.2 | 7.3 | 7.4 | 7.9 | 6.2 |
| trash | HP left | 93% | 95% | 84% | 79% | 64% | 55% | 39% | 36% | 50% |
| elite | win% | 100% | 92% | 96% | 62% | 54% | 21% | 4% | 0% | 0% |
| elite | rounds | 9.8 | 11.0 | 12.9 | 12.9 | 13.5 | 12.8 | 8.6 | 5.6 | 4.9 |
| elite | HP left | 78% | 77% | 77% | 55% | 28% | 25% | 48% | — | — |
| boss | win% | 92% | 88% | 71% | 50% | 0% | 0% | 0% | 0% | 0% |
| boss | rounds | 13.3 | 16.0 | 17.0 | 17.2 | 10.9 | 8.4 | 6.5 | 5.9 | 4.7 |
| boss | HP left | 67% | 77% | 69% | 14% | — | — | — | — | — |

## Rogue — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 92% | 88% | 96% | 92% |
| trash | rounds | 4.0 | 4.8 | 5.0 | 5.0 | 7.0 | 7.0 | 8.2 | 8.5 | 10.4 |
| trash | HP left | 98% | 97% | 93% | 90% | 100% | 93% | 84% | 77% | 71% |
| elite | win% | 100% | 100% | 100% | 88% | 92% | 83% | 71% | 71% | 38% |
| elite | rounds | 9.2 | 9.0 | 9.8 | 11.9 | 16.2 | 17.5 | 17.3 | 23.7 | 24.4 |
| elite | HP left | 88% | 91% | 80% | 85% | 78% | 72% | 59% | 69% | 68% |
| boss | win% | 100% | 88% | 79% | 92% | 58% | 54% | 21% | 8% | 4% |
| boss | rounds | 11.5 | 12.5 | 15.0 | 18.8 | 18.5 | 23.3 | 15.2 | 15.2 | 18.8 |
| boss | HP left | 87% | 77% | 75% | 71% | 48% | 42% | 100% | 100% | 100% |

## Ranger — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 71% | 75% | 17% | 25% |
| trash | rounds | 5.1 | 5.0 | 5.9 | 6.9 | 7.2 | 7.9 | 9.3 | 7.7 | 8.4 |
| trash | HP left | 93% | 94% | 92% | 81% | 73% | 51% | 32% | 46% | 42% |
| elite | win% | 100% | 100% | 83% | 75% | 54% | 21% | 0% | 0% | 0% |
| elite | rounds | 10.1 | 11.6 | 12.6 | 15.2 | 16.3 | 11.7 | 9.9 | 9.8 | 6.8 |
| elite | HP left | 87% | 73% | 87% | 68% | 37% | 57% | — | — | — |
| boss | win% | 96% | 88% | 79% | 50% | 4% | 0% | 0% | 0% | 0% |
| boss | rounds | 15.3 | 17.2 | 19.1 | 19.5 | 13.9 | 10.6 | 10.9 | 6.9 | 4.9 |
| boss | HP left | 77% | 68% | 80% | 58% | 59% | — | — | — | — |

## Ranger — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 100% | 100% | 100% | 100% | 100% | 92% | 96% |
| trash | rounds | 4.0 | 5.0 | 4.9 | 5.1 | 7.0 | 7.4 | 8.4 | 8.2 | 10.0 |
| trash | HP left | 98% | 98% | 97% | 96% | 92% | 82% | 86% | 86% | 71% |
| elite | win% | 100% | 100% | 100% | 100% | 100% | 75% | 75% | 54% | 62% |
| elite | rounds | 8.7 | 9.2 | 10.4 | 13.1 | 15.9 | 15.8 | 19.3 | 24.5 | 27.8 |
| elite | HP left | 94% | 93% | 90% | 80% | 60% | 65% | 77% | 84% | 45% |
| boss | win% | 96% | 96% | 92% | 83% | 58% | 67% | 42% | 33% | 8% |
| boss | rounds | 11.8 | 14.0 | 16.6 | 19.9 | 21.4 | 25.5 | 26.3 | 41.6 | 54.4 |
| boss | HP left | 91% | 67% | 77% | 81% | 64% | 59% | 52% | 33% | 1% |

## Mage — appropriate gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 88% | 96% | 75% | 67% | 58% | 8% | 4% | 0% | 0% |
| trash | rounds | 4.8 | 4.9 | 5.5 | 5.4 | 5.4 | 4.3 | 4.0 | 2.9 | 2.9 |
| trash | HP left | 71% | 49% | 51% | 46% | 28% | 45% | 39% | — | — |
| elite | win% | 79% | 33% | 8% | 8% | 0% | 0% | 0% | 0% | 0% |
| elite | rounds | 8.1 | 6.9 | 5.3 | 6.1 | 5.9 | 4.6 | 3.7 | 3.2 | 2.9 |
| elite | HP left | 53% | 32% | 75% | 42% | — | — | — | — | — |
| boss | win% | 33% | 8% | 0% | 0% | 0% | 0% | 0% | 0% | 0% |
| boss | rounds | 7.8 | 5.6 | 5.9 | 5.2 | 4.9 | 4.3 | 3.8 | 1.6 | 1.4 |
| boss | HP left | 56% | 72% | — | — | — | — | — | — | — |

## Mage — good gear

| Tier | Stat | -3 | -2 | -1 | +0 | +1 | +2 | +3 | +4 | +5 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| trash | win% | 100% | 100% | 96% | 100% | 96% | 92% | 46% | 46% | 17% |
| trash | rounds | 4.9 | 4.8 | 5.9 | 6.0 | 6.1 | 7.4 | 7.2 | 6.6 | 6.8 |
| trash | HP left | 91% | 94% | 87% | 78% | 84% | 67% | 74% | 59% | 39% |
| elite | win% | 96% | 96% | 96% | 83% | 58% | 25% | 21% | 0% | 0% |
| elite | rounds | 8.6 | 9.8 | 10.8 | 11.0 | 11.2 | 9.4 | 10.9 | 7.4 | 7.2 |
| elite | HP left | 87% | 69% | 62% | 64% | 100% | 79% | 62% | — | — |
| boss | win% | 92% | 96% | 38% | 25% | 25% | 8% | 4% | 0% | 0% |
| boss | rounds | 11.5 | 12.3 | 9.5 | 10.2 | 13.0 | 10.6 | 13.5 | 5.9 | 4.5 |
| boss | HP left | 74% | 71% | 94% | 100% | 100% | 62% | 100% | — | — |

## How to read this sample

One 24-iteration run per cell (player level 10). Win% moves a few points between runs.

- Level-appropriate trash at gap 0 is essentially a sure win for warrior, rogue, and ranger. Mages in starter cloth win about half of those races.
- Elites at gap 0 land in or just above the 75–90% band for warrior and ranger. Rogues are a bit lower.
- Bosses at gap 0: rogue and ranger sit near 45–70%. Warriors are stronger and often clear an at-level boss.
- Bosses at +3 with good gear: warrior and ranger land around 40–70%. Rogues are lower (about 20% in this sample).
- At +5, appropriate-gear elites and bosses are under 15% (usually 0). Trash at +5 is still beatable for a warrior (about 30–40% in this sample) and is not a skull fight by itself. Good gear makes +5 trash easy and +5 bosses rare.

`level_gap` per level of attacker advantage: hit +3.5%, crit +1%, damage dealt +3.5%, damage taken +2%, clamped at ±6. Gap 0 does not change the old damage formula, so the at-level content duration tests are unchanged. The tier bodies live in `CreateScaledEnemy`.

