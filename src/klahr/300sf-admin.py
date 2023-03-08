#!/usr/bin/env python

import argparse
import httpx
import json
import os
import sys
import tabulate
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
            file.close()

    def get(self, key):
        try:
            return self.data[key]
        except:
            err_print(f'config key "{key}" not found')
            return None


def get_access_token():
    session_filename = os.path.join(os.path.expanduser('~'), '.config/300sf/session.json')
    with open(session_filename, 'r') as file:
        data = json.load(file)
        return data['access_token']


def http_get(path, headers):
    global config
    if config is not None:
        headers['X-Api-Key'] = str(config.get('api-key'))
        access_token = get_access_token()
        if access_token is not None:
            headers['X-Access-Token'] = access_token
        r = httpx.get(f'{config.get("api-endpoint")}{path}', headers=headers)
        return r
    return None


def http_post(path, headers, body):
    global config
    if config is not None:
        headers['X-Api-Key'] = str(config.get('api-key'))
        access_token = get_access_token()
        if access_token is not None:
            headers['X-Access-Token'] = access_token
        r = httpx.post(f'{config.get("api-endpoint")}{path}', headers=headers, content=json.dumps(body))
        return r
    return None


def version():
    print(__version__)


def named_record(insert, endpoint):
    if insert is not None:
        r = http_post(endpoint, {}, {
            'name': {
                'en': insert
            }
        })
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
    else:
        r = http_get(f'{endpoint}/list', {})
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
        else:
            data = r.json()
            t = []
            for gender in data['result']:
                t.append([gender['uuid'], gender['name']['en'], gender['name']['sv'], gender['name']['la']])
            print(tabulate.tabulate(t, headers=['uuid', 'en', 'sv', 'la'], tablefmt='plain'))


def gender(insert):
    named_record(insert, '/v1/p/genders')
    return


def order(insert):
    named_record(insert, '/v1/p/orders')
    return


def age(insert):
    named_record(insert, '/v1/p/ages')
    return


def activity(insert):
    named_record(insert, '/v1/p/activities')
    return


def method(insert):
    named_record(insert, '/v1/p/methods')
    return


def species(insert, order):
    if insert is not None:
        r = http_post('/v1/p/species', {}, {
            'name': {
                'en': insert
            },
            'order_uuid': order
        })
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
    else:
        r = http_get('/v1/p/species/list', {})
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
        else:
            data = r.json()
            t = []
            for gender in data['result']:
                t.append([gender['uuid'], gender['name']['en'], gender['name']['sv'], gender['name']['la']])
            print(tabulate.tabulate(t, headers=['uuid', 'en', 'sv', 'la'], tablefmt='plain'))


def translation(update, en, la, sv):
    if update is not None:
        r = http_post('/v1/p/translations/update', {}, {
            'uuid': update,
            'en': en,
            'sv': sv,
            'la': la
        });
        if r is not None and r.status_code != 204:
            response = r.json()
            err_print(f'err: {response["message"]}')
    else:
        r = http_get('/v1/p/translations/list', {})
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
        else:
            data = r.json()
            t = []
            for translation in data['translations']:
                t.append([translation['uuid'], translation['en'], translation['sv'], translation['la']])
            print(tabulate.tabulate(t, headers=['uuid', 'en', 'sv', 'la'], tablefmt='plain'))


def main():
    global config
    config = Config(os.path.join(os.path.expanduser('~'), '.config/300sf/config.yml'))

    parser = argparse.ArgumentParser(description='300sf-admin arguments and options')
    subparsers = parser.add_subparsers(dest='subparser')

    gender_parser = subparsers.add_parser('gender', help='Gender.')
    gender_parser.add_argument('--insert', dest='insert', help='Add gender.')

    age_parser = subparsers.add_parser('age', help='Gender.')
    age_parser.add_argument('--insert', dest='insert', help='Add age.')

    order_parser = subparsers.add_parser('order', help='Order.')
    order_parser.add_argument('--insert', dest='insert', help='Add order.')

    activity_parser = subparsers.add_parser('activity', help='Activity.')
    activity_parser.add_argument('--insert', dest='insert', help='Add activity.')

    method_parser = subparsers.add_parser('method', help='Method.')
    method_parser.add_argument('--insert', dest='insert', help='Add method.')

    species_parser = subparsers.add_parser('species', help='Species.')
    species_parser.add_argument('--insert', dest='insert', help='Add species.')
    species_parser.add_argument('--order', dest='order', help='Order')

    translation_parser = subparsers.add_parser('translation', help='Translation.')
    translation_parser.add_argument('--update', dest='update', help='Update translation uuid.')
    translation_parser.add_argument('--en', dest='en', help='English.')
    translation_parser.add_argument('--la', dest='la', help='Swedish.')
    translation_parser.add_argument('--sv', dest='sv', help='Latin.')

    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    kwargs = vars(parser.parse_args())
    globals()[kwargs.pop('subparser')](**kwargs)


if __name__ == '__main__':
    main()
