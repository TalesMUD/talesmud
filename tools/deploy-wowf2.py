#!/usr/bin/env python3
"""Run one verified WOWF deployment stage on the administrative VPS."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import signal
import sqlite3
import subprocess
import time
import sys
from datetime import datetime, timezone
from urllib.request import urlopen

ROOT = Path('/home/atla/dev/talesmud')
PACK = Path('/home/atla/dev/talesmud-rpg-1')
STATE = ROOT / 'backups/wowf2-deploy-state.json'


def run(args, cwd=ROOT):
    return subprocess.check_output(args, cwd=cwd, text=True, stderr=subprocess.PIPE).strip()


def digest(path):
    with open(path, 'rb') as source:
        return hashlib.file_digest(source, 'sha256').hexdigest()


def tree_digest(directory):
    hashes = hashlib.sha256()
    for path in sorted(directory.rglob('*')):
        if path.is_file():
            hashes.update(str(path.relative_to(directory)).encode())
            hashes.update(digest(path).encode())
    return hashes.hexdigest()


def prop(name):
    return run(['systemctl', 'show', 'talesmud.service', f'--property={name}', '--value'])


def save(state):
    temp = STATE.with_suffix('.tmp')
    temp.write_text(json.dumps(state, indent=2) + '\n')
    temp.chmod(0o600)
    os.replace(temp, STATE)


def gate():
    counts = {str(port): 0 for port in (8010, 2222, 2223)}
    for line in run(['ss', '-Htn', 'state', 'established']).splitlines():
        fields = line.split()
        if len(fields) >= 4:
            port = fields[-2].rsplit(':', 1)[-1]
            if port in counts:
                counts[port] += 1
    with sqlite3.connect(f'file:{ROOT}/talesmud.db?mode=ro', uri=True) as source:
        online, total = source.execute("select coalesce(sum(case when json_extract(data,'$.isOnline') or json_extract(data,'$.IsOnline') then 1 else 0 end),0),count(*) from users").fetchone()
    return dict(at=datetime.now(timezone.utc).isoformat(), tcp_established=counts,
                users_online=online, users_total=total,
                clear=not online and not any(counts.values()))


def verify_preserved(state):
    assert all(digest(ROOT / path) == expected for path, expected in state['dirty_hashes'].items())


def backup(args):
    assert not STATE.exists(), 'Existing deploy state requires inspection before another backup'
    pid = int(prop('MainPID'))
    assert pid > 1 and prop('ActiveState') == 'active'
    assert Path(f'/proc/{pid}/exe').resolve() == ROOT / 'bin/tales'
    assert prop('WorkingDirectory') == str(ROOT) and prop('Restart') == 'always'
    assert re.search(r'path=([^ ;]+)', prop('ExecStart')).group(1) == str(ROOT / 'bin/tales')
    assert run(['git', 'branch', '--show-current']) == 'engine-june'
    assert run(['git', 'branch', '--show-current'], PACK) == 'main'
    dirty = run(['git', 'diff', 'HEAD', '--name-only']).splitlines()
    assert all(p == 'public/landing/app/index.html' or p.startswith('import/mvp-rpg-1/assets/images/sprites/portraits/') for p in dirty)
    checked = gate()
    if not checked['clear']:
        print(json.dumps(checked))
        raise SystemExit(42)
    stamp = datetime.now(timezone.utc).strftime('%Y%m%dT%H%M%SZ')
    directory = ROOT / 'backups' / f'wowf2-{stamp}'
    directory.mkdir(mode=0o700)
    state = dict(engine_sha=args.engine_sha, content_sha=args.content_sha,
                 old_sha=run(['git', 'rev-parse', 'HEAD']), old_content_sha=run(['git', 'rev-parse', 'HEAD'], PACK),
                 old_pid=pid, old_binary_sha256=digest(ROOT / 'bin/tales'),
                 dirty_hashes={p: digest(ROOT / p) for p in dirty},
                 directory=str(directory), gate_before_backup=checked)
    assert digest(f'/proc/{pid}/exe') == state['old_binary_sha256']
    target = directory / 'talesmud.db'
    descriptor = os.open(target, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
    os.close(descriptor)
    with sqlite3.connect(f'file:{ROOT}/talesmud.db?mode=ro', uri=True) as source, sqlite3.connect(target) as destination:
        source.backup(destination)
        assert destination.execute('pragma integrity_check').fetchone()[0] == 'ok'
    state['backup'] = dict(path=str(target), size=target.stat().st_size, mode='0o600', integrity='ok')
    rollback = ROOT / 'bin' / f"tales.prev-{state['old_sha'][:7]}-wowf2"
    assert not rollback.exists()
    shutil.copy2(ROOT / 'bin/tales', rollback)
    assert digest(rollback) == state['old_binary_sha256']
    state['rollback_path'] = str(rollback)
    shutil.copytree(ROOT / 'uploads', directory / 'uploads')
    shutil.copytree(ROOT / 'import/mvp-rpg-1', directory / 'import')
    state['uploads_sha256'] = tree_digest(ROOT / 'uploads')
    assert tree_digest(directory / 'uploads') == state['uploads_sha256']
    active = ROOT / 'import/.active'
    state['had_active_marker'] = active.exists()
    if active.exists():
        shutil.copy2(active, directory / 'active')
    state['stage'] = 'backed_up'
    save(state)
    return state


def fetch_target(directory, branch, expected, label):
    for attempt in range(4):
        try:
            subprocess.run(['git', 'fetch', 'origin', branch], cwd=directory,
                           env={**os.environ, 'GIT_TERMINAL_PROMPT': '0'},
                           capture_output=True, timeout=90, check=True)
            break
        except (subprocess.CalledProcessError, subprocess.TimeoutExpired):
            if attempt == 3:
                raise RuntimeError(f'{label} fetch failed after 4 attempts; service and binary unchanged') from None
            delay = (2, 5, 10)[attempt]
            print(f'{label} fetch attempt {attempt + 1}/4 failed; retrying in {delay}s', file=sys.stderr)
            time.sleep(delay)
    assert run(['git', 'rev-parse', 'FETCH_HEAD'], directory) == expected, 'Remote moved; review required'


def build(state):
    assert state['stage'] == 'backed_up'
    verify_preserved(state)
    targets = [(ROOT, 'engine-june', state['old_sha'], state['engine_sha'], 'engine'),
               (PACK, 'main', state['old_content_sha'], state['content_sha'], 'content')]
    for directory, branch, previous, expected, label in targets:
        assert run(['git', 'branch', '--show-current'], directory) == branch
        assert run(['git', 'rev-parse', 'HEAD'], directory) in (previous, expected), 'Checkout changed; review required'
        fetch_target(directory, branch, expected, label)
        run(['git', 'merge-base', '--is-ancestor', previous, expected], directory)
        incoming = set(run(['git', 'diff', '--name-only', previous, expected], directory).splitlines())
        if directory == ROOT:
            assert not incoming.intersection(state['dirty_hashes'])
        else:
            assert not run(['git', 'diff', 'HEAD', '--name-only'], PACK)
    for directory, _, _, expected, _ in targets:
        run(['git', 'merge', '--ff-only', expected], directory)
    staging = Path(state['directory']) / 'pack'
    staging.mkdir(exist_ok=True)
    archive = Path(state['directory']) / 'pack.tar'
    with archive.open('wb') as output:
        subprocess.run(['git', 'archive', state['content_sha']], cwd=PACK, stdout=output, check=True)
    run(['tar', '-xf', str(archive), '-C', str(staging)])
    exclusions = ['--exclude=/' + p.removeprefix('import/mvp-rpg-1/') for p in state['dirty_hashes'] if p.startswith('import/mvp-rpg-1/')]
    run(['rsync', '-a', *exclusions, str(staging) + '/', str(ROOT / 'import/mvp-rpg-1') + '/'])
    verify_preserved(state)
    run(['rsync', '-a', '--delete', 'public/mud-client/public/', 'pkg/webuiplay/dist/'])
    page = (ROOT / 'pkg/webuiplay/dist/index.html').read_text()
    assert 'uxtop1' not in page and '?v=wowf2' in page
    go = shutil.which('go') or '/usr/local/go/bin/go'
    version = run(['env', 'GOTOOLCHAIN=go1.26.9', go, 'version'])
    assert 'go1.26.9' in version
    log = Path(state['directory']) / 'build.log'
    with log.open('wb') as output:
        result = subprocess.run(['env', 'GOTOOLCHAIN=go1.26.9', 'CGO_ENABLED=1', 'GOAMD64=v1', go, 'build', '-o', 'bin/tales.next', './cmd/tales'], cwd=ROOT, stdout=output, stderr=output)
    assert result.returncode == 0, f'Build failed; inspect {log}'
    state['build'] = dict(toolchain=version, CGO_ENABLED='1', GOAMD64='v1', sha256=digest(ROOT / 'bin/tales.next'),
                          assets={name: digest(ROOT / 'public/mud-client/public' / name) for name in ['index.html', 'bundle.js', 'extra.css', 'global.css', 'themes.css']})
    state['stage'] = 'built'
    save(state)
    return state


def restart(state):
    assert state['stage'] == 'built'
    verify_preserved(state)
    assert run(['git', 'rev-parse', 'HEAD']) == state['engine_sha']
    assert digest(ROOT / 'bin/tales.next') == state['build']['sha256']
    pid = int(prop('MainPID'))
    assert pid == state['old_pid'] and prop('ActiveState') == 'active'
    assert digest(f'/proc/{pid}/exe') == state['old_binary_sha256'] == digest(ROOT / 'bin/tales')
    checked = gate()
    state['gate_before_restart'] = checked
    save(state)
    if not checked['clear']:
        print(json.dumps(checked))
        raise SystemExit(42)
    directory = Path(state['directory'])
    hold = directory / 'hold'
    hold.touch(mode=0o600, exist_ok=False)
    wrapper = ROOT / 'bin/tales.wowf-guard'
    wrapper.write_text(f'#!/bin/sh\nwhile test -e {hold}; do sleep 0.25; done\nexec {ROOT}/bin/tales "$@"\n')
    wrapper.chmod(0o755)
    state['hold_path'] = str(hold)
    state['stage'] = 'guarded'
    save(state)
    imported = False
    try:
        assert int(prop('MainPID')) == pid
        os.replace(wrapper, ROOT / 'bin/tales')
        os.kill(pid, signal.SIGTERM)
        state['restart'] = dict(signaled_pid=pid, at=datetime.now(timezone.utc).isoformat(), signal='SIGTERM')
        save(state)
        for _ in range(120):
            if not Path(f'/proc/{pid}').exists():
                break
            time.sleep(0.25)
        else:
            raise RuntimeError('Old talesmud PID did not exit; import was not started')
        stopped_backup = directory / 'talesmud-stopped.db'
        descriptor = os.open(stopped_backup, os.O_CREAT | os.O_EXCL | os.O_WRONLY, 0o600)
        os.close(descriptor)
        with sqlite3.connect(f'file:{ROOT}/talesmud.db?mode=ro', uri=True) as source, sqlite3.connect(stopped_backup) as destination:
            source.backup(destination)
            assert destination.execute('pragma integrity_check').fetchone()[0] == 'ok'
        state['stopped_backup'] = dict(path=str(stopped_backup), size=stopped_backup.stat().st_size, mode='0o600', integrity='ok')
        save(state)
        imported = True
        log = directory / 'import.log'
        env = {**os.environ, 'SQLITE_PATH': str(ROOT / 'talesmud.db')}
        with log.open('wb') as output:
            result = subprocess.run([str(ROOT / 'bin/tales.next'), '--import', 'mvp-rpg-1'], cwd=ROOT, env=env, stdout=output, stderr=output)
        log.chmod(0o600)
        assert result.returncode == 0, f'Import failed; inspect {log}'
        assert (ROOT / 'import/.active').read_text().strip() == 'mvp-rpg-1'
        assert b'Import completed successfully!' in log.read_bytes()
        run(['rsync', '-a', '--delete', str(directory / 'uploads') + '/', str(ROOT / 'uploads') + '/'])
        assert tree_digest(ROOT / 'uploads') == state['uploads_sha256']
        verify_preserved(state)
        state['import'] = dict(exit=0, log=str(log), restored_uploads=True)
        state['stage'] = 'imported'
        save(state)
        os.replace(ROOT / 'bin/tales.next', ROOT / 'bin/tales')
    except BaseException:
        if imported:
            with sqlite3.connect(state['stopped_backup']['path']) as source, sqlite3.connect(ROOT / 'talesmud.db') as destination:
                source.backup(destination)
                assert destination.execute('pragma integrity_check').fetchone()[0] == 'ok'
            run(['rsync', '-a', '--delete', str(directory / 'import') + '/', str(ROOT / 'import/mvp-rpg-1') + '/'])
            run(['rsync', '-a', '--delete', str(directory / 'uploads') + '/', str(ROOT / 'uploads') + '/'])
            if state['had_active_marker']:
                shutil.copy2(directory / 'active', ROOT / 'import/.active')
            else:
                (ROOT / 'import/.active').unlink(missing_ok=True)
        restored = ROOT / 'bin/tales.wowf-restore'
        shutil.copy2(state['rollback_path'], restored)
        os.replace(restored, ROOT / 'bin/tales')
        state['stage'] = 'rolled_back'
        save(state)
        raise
    finally:
        with (ROOT / 'bin/tales').open('rb') as binary:
            installed_elf = binary.read(4) == b'\x7fELF'
        if installed_elf:
            hold.unlink(missing_ok=True)
    for _ in range(60):
        current = int(prop('MainPID'))
        if current > 1 and current != pid and prop('ActiveState') == 'active':
            try:
                with urlopen('http://127.0.0.1:8010/health', timeout=2) as response:
                    assert response.status == 200
                if digest(f'/proc/{current}/exe') == state['build']['sha256']:
                    state['restart'].update(new_pid=current, health=200, running_sha256=digest(f'/proc/{current}/exe'))
                    break
            except (OSError, AssertionError):
                pass
        time.sleep(1)
    else:
        state['stage'] = 'verification_failed'
        save(state)
        raise RuntimeError('New talesmud service failed verification')
    verify_preserved(state)
    state['stage'] = 'deployed'
    save(state)
    return state


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('stage', choices=['gate', 'backup', 'build', 'restart'])
    parser.add_argument('--engine-sha')
    parser.add_argument('--content-sha')
    args = parser.parse_args()
    if args.stage == 'gate':
        print(json.dumps(gate()))
        return
    if args.stage == 'backup':
        assert all(re.fullmatch('[0-9a-f]{40}', value or '') for value in [args.engine_sha, args.content_sha])
        state = backup(args)
    else:
        state = json.loads(STATE.read_text())
        state = {'build': build, 'restart': restart}[args.stage](state)
    print(json.dumps(state))


if __name__ == '__main__':
    main()
