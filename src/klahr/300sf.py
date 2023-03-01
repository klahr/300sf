import argparse
import httpx
import os
import sys
import yaml

__version__ = '0.1.0'
config = None


def err_print(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)


class Config:
    data = {}

    def __init__(self, filename):
        with open(filename) as file:
            stream = file.read()
            self.data = yaml.safe_load(stream)

    def get(self, key):
        try:
            return self.data[key]
        except:
            err_print(f'config key "{key}" not found')
            return None


def http_get(path):
    global config
    headers = {'X-Api-Key': str(config.get('api-key'))}
    r = httpx.get(f'{config.get("api-endpoint")}{path}', headers=headers)
    return r


def version():
    print(__version__)


def signup(display_name, password, email):
    r = http_get('/v1/test')
    print(r.status_code)
    print(r.text)


def main():
    global config
    config = Config('/home/klarre/.config/300sf/config.yml')

    parser = argparse.ArgumentParser(description='300sf arguments and options')
    subparsers = parser.add_subparsers(dest='subparser')

    subparsers.add_parser('version')

    signup_parser = subparsers.add_parser('signup', help='Sign up for a new account.')
    signup_parser.add_argument('--display-name', dest='display_name', help='Public display name.')
    signup_parser.add_argument('--password', dest='password', help='Log in password.')
    signup_parser.add_argument('--email', dest='email', help='Email address.')

    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    kwargs = vars(parser.parse_args())
    globals()[kwargs.pop('subparser')](**kwargs)


if __name__ == '__main__':
    main()
