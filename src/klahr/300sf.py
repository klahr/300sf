#!/usr/bin/env python

import argparse
import base64
import httpx
import datetime
import json
import os
import sys
import tabulate
import yaml

__version__ = '0.1.0'


def err_print(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)


class Config:
    data = {}

    def open(self, filename):
        with open(filename) as file:
            stream = file.read()
            self.data = yaml.safe_load(stream)
            file.close()

    def get(self, key):
        try:
            return self.data[key]
        except:
            err_print(f'config key "{key}" not found')
            return None


config = Config()


def get_access_token():
    session_filename = os.path.join(os.path.expanduser('~'), '.config/300sf/session.json')
    if os.path.exists(session_filename):
        with open(session_filename, 'r') as file:
            data = json.load(file)
            return data['access_token']
    return None


def http_get(path, headers):
    global config
 
    headers['X-Api-Key'] = str(config.get('api-key'))
    access_token = get_access_token()
    if access_token is not None:
        headers['X-Access-Token'] = access_token
    r = httpx.get(f'{config.get("api-endpoint")}{path}', headers=headers)
    return r


def http_post(path, headers, body):
    global config

    headers['X-Api-Key'] = str(config.get('api-key'))
    access_token = get_access_token()
    if access_token is not None:
        headers['X-Access-Token'] = access_token
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
    if r is not None and r.status_code != 200:
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
    if password == None and not sys.stdin.isatty():
        password = sys.stdin.read()

    if password == None:
        err_print('err: no password (--password) provided.')
        return

    r = http_post('/v1/auth', {
        'Authorization': 'Basic ' + base64.b64encode((email + ':' + password).encode()).decode(),
    }, {})
    if r is not None:
        if r.status_code == 200:
            session_filename = os.path.join(os.path.expanduser('~'), '.config/300sf/session.json')
            with open(session_filename, 'w') as file:
                file.write(json.dumps(r.json(), indent=4))
                file.close()
        else:
            response = r.json()
            err_print(f'err: {response["message"]}')
    else:
        err_print('err: internal error')


def logout():
    session_filename = os.path.join(os.path.expanduser('~'), '.config/300sf/session.json')
    if os.path.exists(session_filename):
        os.remove(session_filename)
    else:
        err_print('err: already logged out')


# create is name and revoke is prefix.
def apikey(create, revoke):
    if create is not None:
        r = http_post('/v1/api_keys', {}, {
            'name': str(create or ''),
        })
        if r is not None:
            if r.status_code == 200:
                print(r.json()['api-key'])
            else:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif revoke is not None:
        r = http_post(f'/v1/api_keys/{revoke}/revoke', {}, {})
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    else:
        r = http_get('/v1/api_keys/list', {})
        if r is not None:
            if r.status_code == 200:
                data = r.json()
                t = []
                for api_key in data['api_keys']:
                    revoked_at = datetime.datetime.fromisoformat(api_key['revoked_at'].replace('Z', '+00:00'))
                    api_key_revoked_at = ''
                    if revoked_at.year > 1:
                        api_key_revoked_at = revoked_at.isoformat()
                    t.append([api_key['name'], api_key['prefix'], api_key['scopes'], api_key['soft_limit'], api_key['hard_limit'], api_key_revoked_at])
                print(tabulate.tabulate(t, headers=['name', 'prefix', 'scopes', 'soft limit', 'hard limit', 'revoked at'], tablefmt='plain'))
            else:
                err_print(f'err: {r.json()["message"]}')


def listtemplate(create, visibility):
    global config
    if create is not None:
        r = http_post('/v1/list_templates', {}, {
            'name': {
                config.get('language'): create
            },
            'visibility': int(visibility)
        })
        if r is not None:
            if r.status_code == 200:
                print(r.json()['uuid'])
            else:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    else:
        r = http_get('/v1/list_templates/list', {})
        if r is not None:
            if r.status_code == 200:
                data = r.json()
                t = []
                for template in data['result']:
                    t.append([template['uuid'], template['name'][config.get('language')], template['visibility']])
                print(tabulate.tabulate(t, headers=['uuid', f'name ({config.get("language")})', 'visibility'], tablefmt='plain'))


def species(insert, order):
    global config

    if insert is not None:
        r = http_post('/v1/p/species', {}, {
            'name': {
                config.get('language'): insert
            },
            'order_uuid': order
        })
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
    else:
        r = http_get('/v1/species/list', {})
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
        else:
            data = r.json()
            print(data)
            t = []
            for species in data['result']:
                t.append([species['uuid'], species['name'][config.get('language')], species['order']['name'][config.get('language')]])
            print(tabulate.tabulate(t, headers=['uuid', f'name ({config.get("language")})', f'order ({config.get("language")})'], tablefmt='plain'))


def main():
    global config
    config.open(os.path.join(os.path.expanduser('~'), '.config/300sf/config.yml'))

    parser = argparse.ArgumentParser(description='300sf arguments and options')
    subparsers = parser.add_subparsers(dest='subparser')

    subparsers.add_parser('version')

    signup_parser = subparsers.add_parser('signup', help='Sign up for a new account.')
    signup_parser.add_argument('--display-name', dest='display_name', required=True, help='Public display name.')
    signup_parser.add_argument('--password', dest='password', required=True, help='Log in password.')
    signup_parser.add_argument('--email', dest='email', required=True, help='Email address.')
    signup_parser.add_argument('--accepted-terms', action='store_true', dest='accepted_terms', required=True, help='Terms accepted.')

    login_parser = subparsers.add_parser('login', help='Log in.')
    login_parser.add_argument('--email', dest='email', required=True, help='Email address.')
    login_parser.add_argument('--password', dest='password', help='Log in password')

    subparsers.add_parser('logout', help='Log out.')

    subparsers.add_parser('test')

    api_key_parser = subparsers.add_parser('apikey', help='API key.')
    api_key_parser.add_argument('--create', dest='create', help='Create new API key.')
    api_key_parser.add_argument('--revoke', dest='revoke', help='Revoke API key.')
#    api_key_parser.add_argument('--scopes', dest='scopes', help='Scopes')

    list_template_parser = subparsers.add_parser('listtemplate', help='List template.')
    list_template_parser.add_argument('--create', dest='create', help='Create new list template.')
    list_template_parser.add_argument('--visibility', dest='visibility', choices=[0, 1, 2], help='Visibility.')

    species_parser = subparsers.add_parser('species', help='Species.')
    species_parser.add_argument('--insert', dest='insert', help='Add species.')
    species_parser.add_argument('--order', dest='order', help='Order.')

    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    kwargs = vars(parser.parse_args())
    globals()[kwargs.pop('subparser')](**kwargs)


if __name__ == '__main__':
    main()
