import contextlib
import importlib.util
import io
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch


spec = importlib.util.spec_from_file_location('deploy', Path(__file__).with_name('deploy-wowf2.py'))
deploy = importlib.util.module_from_spec(spec)
spec.loader.exec_module(deploy)


class BuildSafetyTest(unittest.TestCase):
    def git(self, directory, *args):
        return subprocess.check_output(['git', *args], cwd=directory, text=True, stderr=subprocess.DEVNULL).strip()

    def repository(self, name, branch):
        directory = self.base / name
        directory.mkdir()
        self.git(directory, 'init', '-b', branch)
        self.git(directory, 'config', 'user.email', 'fixture@example.invalid')
        self.git(directory, 'config', 'user.name', 'Deploy fixture')
        (directory / 'version').write_text('old')
        self.git(directory, 'add', '.')
        self.git(directory, 'commit', '-m', 'test: [grokbot] old deployment fixture')
        previous = self.git(directory, 'rev-parse', 'HEAD')
        (directory / 'version').write_text('target')
        self.git(directory, 'commit', '-am', 'test: [grokbot] target deployment fixture')
        target = self.git(directory, 'rev-parse', 'HEAD')
        remote = self.base / (name + '.git')
        self.git(directory, 'clone', '--bare', str(directory), str(remote))
        self.git(directory, 'remote', 'add', 'origin', str(remote))
        self.git(directory, 'reset', '--hard', previous)
        return directory, previous, target

    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.base = Path(self.temp.name)
        self.root, old, target = self.repository('engine', 'engine-june')
        self.pack, old_pack, target_pack = self.repository('content', 'main')
        directory = self.base / 'backup'
        directory.mkdir()
        (self.root / 'bin').mkdir()
        (self.root / 'bin/tales').write_bytes(b'old running binary')
        (self.root / 'import/mvp-rpg-1').mkdir(parents=True)
        (self.root / 'pkg/webuiplay/dist').mkdir(parents=True)
        public = self.root / 'public/mud-client/public'
        public.mkdir(parents=True)
        for name in ['index.html', 'bundle.js', 'extra.css', 'global.css', 'themes.css']:
            (public / name).write_text('?v=wowf2')
        self.go = self.base / 'go'
        self.go.write_text('#!/bin/sh\nif [ "$1" = version ]; then\n  echo "go version go1.26.9 linux/amd64"\nelse\n  printf "\\177ELFverified" > bin/tales.next\nfi\n')
        self.go.chmod(0o755)
        self.state = dict(stage='backed_up', old_sha=old, engine_sha=target,
                          old_content_sha=old_pack, content_sha=target_pack,
                          dirty_hashes={}, directory=str(directory))
        for attr, value in [('ROOT', self.root), ('PACK', self.pack),
                            ('STATE', directory / 'state.json')]:
            replace = patch.object(deploy, attr, value)
            replace.start()
            self.addCleanup(replace.stop)
        go = patch.object(deploy.shutil, 'which', return_value=str(self.go))
        go.start()
        self.addCleanup(go.stop)
        default_root = patch.object(deploy.run, '__defaults__', (self.root,))
        default_root.start()
        self.addCleanup(default_root.stop)
        self.real_run = subprocess.run

    def flaky_content(self, failures, timeout=False):
        remaining = failures

        def run(args, **kwargs):
            nonlocal remaining
            if args[:2] == ['git', 'fetch'] and kwargs['cwd'] == self.pack and remaining:
                remaining -= 1
                if timeout:
                    raise subprocess.TimeoutExpired(args, 90, stderr=b'credential-must-not-escape')
                raise subprocess.CalledProcessError(128, args, stderr=b'credential-must-not-escape')
            return self.real_run(args, **kwargs)

        return patch.object(deploy.subprocess, 'run', side_effect=run)

    def assert_unchanged(self):
        self.assertEqual(self.git(self.root, 'rev-parse', 'HEAD'), self.state['old_sha'])
        self.assertEqual(self.git(self.pack, 'rev-parse', 'HEAD'), self.state['old_content_sha'])
        self.assertEqual((self.root / 'bin/tales').read_bytes(), b'old running binary')
        self.assertEqual(self.state['stage'], 'backed_up')

    def test_transient_content_failure_recovers(self):
        errors = io.StringIO()
        with self.flaky_content(2), patch.object(deploy.time, 'sleep'), contextlib.redirect_stderr(errors):
            deploy.build(self.state)
        self.assertEqual(self.state['stage'], 'built')
        self.assertEqual(self.git(self.root, 'rev-parse', 'HEAD'), self.state['engine_sha'])
        self.assertEqual(self.git(self.pack, 'rev-parse', 'HEAD'), self.state['content_sha'])
        self.assertNotIn('credential-must-not-escape', errors.getvalue())
        self.assertEqual((self.root / 'bin/tales').read_bytes(), b'old running binary')

    def test_exhausted_fetch_leaves_both_checkouts_and_binary_unchanged(self):
        errors = io.StringIO()
        with self.flaky_content(4), patch.object(deploy.time, 'sleep'), contextlib.redirect_stderr(errors):
            with self.assertRaises(RuntimeError) as failure:
                deploy.build(self.state)
        self.assertNotIn('credential-must-not-escape', str(failure.exception) + errors.getvalue())
        self.assert_unchanged()

    def test_fetch_timeout_fails_safe(self):
        with self.flaky_content(4, timeout=True), patch.object(deploy.time, 'sleep'), contextlib.redirect_stderr(io.StringIO()):
            with self.assertRaises(RuntimeError):
                deploy.build(self.state)
        self.assert_unchanged()

    def test_remote_moved_rejects_before_merging(self):
        self.state['content_sha'] = self.state['old_content_sha']
        with self.assertRaisesRegex(AssertionError, 'Remote moved'):
            deploy.build(self.state)
        self.assert_unchanged()

    def test_resumes_after_engine_already_advanced_and_staging_exists(self):
        self.git(self.root, 'merge', '--ff-only', self.state['engine_sha'])
        (Path(self.state['directory']) / 'pack').mkdir()
        deploy.build(self.state)
        self.assertEqual(self.state['stage'], 'built')
        self.assertEqual((self.root / 'import/mvp-rpg-1/version').read_text(), 'target')
        self.assertEqual((self.root / 'bin/tales').read_bytes(), b'old running binary')

    def test_build_failure_can_be_retried(self):
        original = self.go.read_text()
        self.go.write_text('#!/bin/sh\nif [ "$1" = version ]; then echo "go version go1.26.9 linux/amd64"; else exit 1; fi\n')
        with self.assertRaisesRegex(AssertionError, 'Build failed'):
            deploy.build(self.state)
        self.assertEqual(self.state['stage'], 'backed_up')
        self.go.write_text(original)
        deploy.build(self.state)
        self.assertEqual(self.state['stage'], 'built')
        self.assertEqual((self.root / 'bin/tales').read_bytes(), b'old running binary')


if __name__ == '__main__':
    unittest.main()
