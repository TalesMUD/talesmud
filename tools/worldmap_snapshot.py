#!/usr/bin/env python3
"""Export only atlas input fields. SQLite opens read-only; YAML needs PyYAML."""
import argparse
import json
import sqlite3
from pathlib import Path
ap=argparse.ArgumentParser()
sources=ap.add_mutually_exclusive_group(required=True)
sources.add_argument('--database',type=Path)
sources.add_argument('--content-rooms',type=Path)
ap.add_argument('--out',type=Path,required=True)
a=ap.parse_args()
if a.database:
    with sqlite3.connect(a.database.resolve().as_uri()+'?mode=ro',uri=True) as conn:
        records=[json.loads(row[0]) for row in conn.execute('select data from rooms order by id')]
else:
    import yaml
    records=[yaml.safe_load(p.read_text()) for p in sorted(a.content_rooms.glob('*.yaml'))]
keys=['id','name','description','detail','roomType','areaType','area','tags','coords','exits','canBind','actions']
exported=[{key:r[key] for key in keys if key in r} for r in records]
# Art classification needs service action names, never scripts/params/responses.
for r in exported:
    if isinstance(r.get('actions'),list):
        r['actions']=[{'name':action.get('name','')} for action in r['actions'] if isinstance(action,dict)]
a.out.write_text(json.dumps(exported,indent=2)+'\n')
print(f'Exported {len(records)} rooms to {a.out}')
