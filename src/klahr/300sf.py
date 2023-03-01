import argparse
import base64
import httpx
import json
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


def http_post(path, headers, body):
    global config
    headers['X-Api-Key'] = str(config.get('api-key'))
    r = httpx.post(f'{config.get("api-endpoint")}{path}', headers=headers, content=json.dumps(body))
    return r

def version():
    print(__version__)


def signup(display_name, password, email, accepted_terms):
    r = http_post('/v1/users', {}, {
        'password': str(password or ''),
        'display_name': str(display_name or ''),
        'email': str(email or ''),
        'accepted_terms': accepted_terms
    })
    if r.status_code == 200:
        print('success')
    else:
        response = r.json()
        code = response['code']
        if code == -20001:
            err_print('err: terms must be accepted.')
        elif code == -20002:
            err_print('err: email is invalid or already taken.')
        elif code == -20003:
            err_print('err: display name is invalid.')
        elif code == -20004:
            err_print('err: password is invalid.')
        else:
            err_print(f'err: unknown error {response["message"]}')

def login(email, password):
    if email == None:
        err_print('err: no email (--email) provided.')
        return
    elif password == None:
        err_print('err: no password (--password) provided.')
        return

    r = http_post('/v1/auth', {
        'Authorization': 'Basic ' + base64.b64encode((email + ':' + password).encode()).decode(),
    }, {})
    if r.status_code == 200:
        print("success")
        #print(r.json())
    else:
        response = r.json()
        err_print(f'err: {response["message"]}')


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
    signup_parser.add_argument('--accepted-terms', action="store_true", dest='accepted_terms', help='Terms accepted.')

    login_parser = subparsers.add_parser('login', help='Log in.')
    login_parser.add_argument('--email', dest='email', help='Email address.')
    login_parser.add_argument('--password', dest='password', help='Log in password')

    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    kwargs = vars(parser.parse_args())
    globals()[kwargs.pop('subparser')](**kwargs)


if __name__ == '__main__':
    main()
