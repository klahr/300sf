import io
import json
import os
import stat
import sys
import tempfile
import unittest
from contextlib import redirect_stderr, redirect_stdout
from unittest import mock

from klahr import cli


class FakeResponse:
    def __init__(self, status_code, data):
        self.status_code = status_code
        self._data = data

    def json(self):
        return self._data


class CliTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        config_dir = os.path.join(self.tmp.name, "300sf")
        for name, value in {
            "CONFIG_DIR": config_dir,
            "SESSION_FILENAME": os.path.join(config_dir, "session.json"),
        }.items():
            patcher = mock.patch.object(cli, name, value)
            patcher.start()
            self.addCleanup(patcher.stop)
        cli.config.data = {
            "api-endpoint": "http://api.test",
            "client-id": "cid",
            "client-secret": "pfx.secret",
        }

    def write_session(self, token):
        os.makedirs(cli.CONFIG_DIR, exist_ok=True)
        with open(cli.SESSION_FILENAME, "w", encoding="utf8") as file:
            json.dump({"access_token": token}, file)


class TestHelpers(CliTest):
    def test_get_date(self):
        self.assertEqual(cli.get_date("2024-05-12T10:00:00Z"), "2024-05-12T10:00:00+00:00")
        self.assertEqual(cli.get_date("0001-01-01T00:00:00Z"), "")

    def test_config_missing_key(self):
        with redirect_stderr(io.StringIO()) as err:
            self.assertIsNone(cli.config.get("nope"))
        self.assertIn('"nope" not found', err.getvalue())

    def test_access_token(self):
        self.assertIsNone(cli.get_access_token())
        self.write_session("tok")
        self.assertEqual(cli.get_access_token(), "tok")


class TestHttp(CliTest):
    def test_get_sends_bearer_token(self):
        self.write_session("tok")
        with mock.patch("httpx.get", return_value=FakeResponse(200, {})) as get:
            cli.http_get("/v1/users/whoami", {})
        get.assert_called_once()
        self.assertEqual(get.call_args.args[0], "http://api.test/v1/users/whoami")
        self.assertEqual(get.call_args.kwargs["headers"]["Authorization"], "Bearer tok")

    def test_post_without_session_has_no_auth(self):
        with mock.patch("httpx.post", return_value=FakeResponse(200, {})) as post:
            cli.http_post("/v1/tags", {}, {"name": "x"})
        self.assertNotIn("Authorization", post.call_args.kwargs["headers"])
        self.assertEqual(json.loads(post.call_args.kwargs["content"]), {"name": "x"})

    def test_connection_error_exits(self):
        with mock.patch("httpx.get", side_effect=Exception("down")), redirect_stderr(io.StringIO()):
            with self.assertRaises(SystemExit):
                cli.http_get("/health", {})


class TestLogin(CliTest):
    def test_login_from_stdin_writes_private_session(self):
        stdin = io.StringIO("s3cret\n")
        stdin.isatty = lambda: False
        response = FakeResponse(200, {"access_token": "at", "refresh_token": "rt"})
        with mock.patch.object(sys, "stdin", stdin), mock.patch("httpx.post", return_value=response) as post:
            cli.login("me@example.com", None)
        body = json.loads(post.call_args.kwargs["content"])
        self.assertEqual(body["password"], "s3cret")
        self.assertEqual(body["client_id"], "cid")
        self.assertEqual(body["client_secret"], "pfx.secret")
        self.assertEqual(body["grant_type"], "password")
        self.assertEqual(stat.S_IMODE(os.stat(cli.SESSION_FILENAME).st_mode), 0o600)
        self.assertEqual(cli.get_access_token(), "at")

    def test_login_failure_keeps_logged_out(self):
        with mock.patch("httpx.post", return_value=FakeResponse(400, {"message": "invalid"})), \
                redirect_stderr(io.StringIO()) as err:
            cli.login("me@example.com", "wrong")
        self.assertIn("invalid", err.getvalue())
        self.assertFalse(os.path.exists(cli.SESSION_FILENAME))

    def test_logout(self):
        self.write_session("tok")
        cli.logout()
        self.assertFalse(os.path.exists(cli.SESSION_FILENAME))
        with redirect_stderr(io.StringIO()) as err:
            cli.logout()
        self.assertIn("already logged out", err.getvalue())


class TestMain(CliTest):
    def run_main(self, *args):
        os.makedirs(cli.CONFIG_DIR, exist_ok=True)
        with open(os.path.join(cli.CONFIG_DIR, "config.yml"), "w", encoding="utf8") as file:
            file.write("api-endpoint: http://api.test\n")
        with mock.patch.object(sys, "argv", ["300sf", *args]), redirect_stdout(io.StringIO()) as out:
            cli.main()
        return out.getvalue()

    def test_version(self):
        self.assertEqual(self.run_main("version").strip(), cli.__version__)

    def test_tag_list(self):
        response = FakeResponse(200, {"result": [{"uuid": "u1", "name": "garden"}]})
        with mock.patch("httpx.get", return_value=response) as get:
            out = self.run_main("tag")
        self.assertEqual(get.call_args.args[0], "http://api.test/v1/tags/list")
        self.assertIn("garden", out)


if __name__ == "__main__":
    unittest.main()
