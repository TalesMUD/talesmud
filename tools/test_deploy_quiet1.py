from contextlib import closing
import importlib.util
import json
from pathlib import Path
import sqlite3
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('deploy_quiet1', Path(__file__).with_name('deploy-quiet1.py'))
deploy = importlib.util.module_from_spec(spec)
spec.loader.exec_module(deploy)


class OnlineGateTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        root = Path(self.temp.name)
        self.root = root
        with closing(sqlite3.connect(root / 'talesmud.db')) as db, db:
            db.execute('create table users (data text)')
            db.execute('insert into users values (?)', (json.dumps({'isOnline': False}),))
        root_patch = patch.object(deploy, 'ROOT', root)
        root_patch.start()
        self.addCleanup(root_patch.stop)

    def gate(self, rows=''):
        with patch.object(deploy, 'run', return_value=rows):
            return deploy.gate()

    def test_empty_gate_is_clear(self):
        self.assertTrue(self.gate()['clear'])

    def test_each_game_port_holds_deploy(self):
        for port in (8010, 2222, 2223):
            with self.subTest(port=port):
                checked = self.gate(f'0 0 127.0.0.1:{port} 127.0.0.2:12345')
                self.assertFalse(checked['clear'])
                self.assertEqual(checked['tcp_established'][str(port)], 1)

    def test_online_user_holds_deploy_without_tcp(self):
        with closing(sqlite3.connect(self.root / 'talesmud.db')) as db, db:
            db.execute('update users set data=?', (json.dumps({'isOnline': True}),))
        self.assertFalse(self.gate()['clear'])

    def test_admin_ssh_does_not_count_as_a_player(self):
        self.assertTrue(self.gate('0 0 127.0.0.1:22 127.0.0.2:8010')['clear'])


if __name__ == '__main__':
    unittest.main()
