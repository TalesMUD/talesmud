"""Build screenshot contacts and validate the measured Quiet Dark touch controls."""
import json
from pathlib import Path
from PIL import Image, ImageDraw

ROOT = Path('/home/clawd/dev/briefs/quiet-dark-shots')


def luminance(color):
    channels = [int(color[i:i + 2], 16) / 255 for i in (1, 3, 5)]
    linear = [c / 12.92 if c <= 0.04045 else ((c + 0.055) / 1.055) ** 2.4 for c in channels]
    return sum(c * weight for c, weight in zip(linear, (0.2126, 0.7152, 0.0722)))


def contrast(fg, bg):
    high, low = sorted((luminance(fg), luminance(bg)), reverse=True)
    return round((high + 0.05) / (low + 0.05), 3)


pairs = [dict(fg=fg, bg=bg, ratio=contrast(fg, bg))
         for fg in ('#ece6dc', '#b5ada1', '#d9b46a', '#e7796c', '#7cc27a', '#7aa6e0')
         for bg in ('#0d0c0b', '#151312', '#1c1a18', '#25221f')]
pairs.append(dict(fg='#1a1408', bg='#d9b46a', ratio=contrast('#1a1408', '#d9b46a')))
assert min(pair['ratio'] for pair in pairs) >= 4.5
(ROOT / 'contrast.json').write_text(json.dumps(pairs, indent=2) + '\n')

for phase in ('before', 'after'):
    tiles = []
    for client in ('web', 'flutter'):
        for size in ('1024x768', '375x812', '1440x900'):
            directory = ROOT / phase / client / size
            for screen in ('main', 'room', 'character', 'character-sheet', 'enemies', 'combat'):
                file = directory / (screen + '.png')
                if not file.exists():
                    continue
                with Image.open(file) as original:
                    preview = original.convert('RGB')
                    preview.thumbnail((360, 260))
                tile = Image.new('RGB', (380, 294), '#0d0c0b')
                tile.paste(preview, ((380 - preview.width) // 2, 26))
                ImageDraw.Draw(tile).text((10, 8), f'{client} {size} {screen}', fill='#ece6dc')
                tiles.append(tile)
    sheet = Image.new('RGB', (1520, 294 * ((len(tiles) + 3) // 4)), '#0d0c0b')
    for index, tile in enumerate(tiles):
        sheet.paste(tile, (index % 4 * 380, index // 4 * 294))
    sheet.save(ROOT / f'contact-{phase}.png')

measurements = []
for file in sorted((ROOT / 'after/web').glob('*/metrics.json')):
    row = json.loads(file.read_text())
    if row['width'] > 1024:
        continue
    controls = [control for controls in row['geometry'].values() for control in controls]
    failures = [c for c in controls if c['width'] < 43.9 or c['height'] < 43.9]
    measurements.append(dict(width=row['width'], samples=len(controls), failures=failures))
    assert not failures, failures
(ROOT / 'touch-geometry.json').write_text(json.dumps(measurements, indent=2) + '\n')
print(json.dumps(dict(minimum_contrast=min(pair['ratio'] for pair in pairs), touch=measurements)))
