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
from getpass import getpass

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
    try:
        r = httpx.get(f'{config.get("api-endpoint")}{path}', headers=headers, timeout=30.0)
        return r
    except:
        err_print('err: could not connect to server')
        sys.exit(1)


def http_post(path, headers, body):
    global config

    headers['X-Api-Key'] = str(config.get('api-key'))
    access_token = get_access_token()
    if access_token is not None:
        headers['X-Access-Token'] = access_token
    try:
        r = httpx.post(f'{config.get("api-endpoint")}{path}', headers=headers, content=json.dumps(body), timeout=30.0)
        return r
    except:
        err_print('err: could not connect to server')
        sys.exit(1)


def version():
    print(__version__)


def get_date(date):
    d = datetime.datetime.fromisoformat(date.replace('Z', '+00:00'))
    if d.year > 1:
        return d.isoformat()
    return ''


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
        password = getpass()

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
                    api_key_revoked_at = get_date(api_key['revoked_at'])
                    t.append([api_key['name'], api_key['prefix'], api_key['scopes'], api_key['soft_limit'], api_key['hard_limit'], api_key_revoked_at])
                print(tabulate.tabulate(t, headers=['name', 'prefix', 'scopes', 'soft limit', 'hard limit', 'revoked at'], tablefmt='plain'))
            else:
                err_print(f'err: {r.json()["message"]}')


def template(create, description, visibility, add, remove, archive, rename, lang, details):
    global config
    if create is not None:
        v = 0
        if visibility == 'hidden':
            v = 1
        elif visibility == 'public':
            v = 2
        r = http_post('/v1/list_templates', {}, {
            'name': {
                lang: create
            },
            'description': {
                lang: description
            },
            'visibility': v
        })
        if r is not None:
            if r.status_code == 200:
                print(r.json()['uuid'])
            else:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif add is not None:
        r = http_post(f'/v1/list_templates/{add[0]}/add_species', {}, {
            'species_uuid': add[1]
        })
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif remove is not None:
        r = http_post(f'/v1/list_templates/{remove[0]}/remove_species', {}, {
            'species_uuid': remove[1]
        })
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif rename is not None:
        r = http_post(f'/v1/list_templates/{rename[0]}/rename', {}, {
            'name': {
                lang: rename[1]
            }
        })
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif archive:
        r = http_post(f'/v1/list_templates/{archive}/archive', {}, {})
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif details:
        r = http_get(f'/v1/list_templates/{details}', {})
        if r is not None:
            if r.status_code == 200:
                response = r.json()['result']
                v = 'private'
                if response['visibility'] == 1:
                    v = 'hidden'
                elif response['visibility'] == 2:
                    v = 'public'
                print(f'id           {response["uuid"]}')
                print(f'name         {response["name"][lang]}')
                print(f'visibility   {v}')
                print(f'description  {response["description"][lang]}')
                species = response['species']
                if species:
                    for spec in species:
                        print(f'species      {spec["name"][lang]}')
            else:
                err_print(f'err: {r.json()["message"]}')
    else:
        r = http_get('/v1/list_templates/list', {})
        if r is not None:
            if r.status_code == 200:
                data = r.json()
                t = []
                for template in data['result']:
                    visibility = 'public'
                    if template['visibility'] == 0:
                        visibility = 'private'
                    elif template['visibility'] == 1:
                        visibility = 'hidden'
                    species = 0
                    if template['species'] is not None:
                        species = len(template['species'])
                    t.append([template['uuid'], template['name'][lang], visibility, species])
                print(tabulate.tabulate(t, headers=['id', f'name({lang})', 'visibility', 'species'], tablefmt='plain'))
            else:
                err_print(f'err: {r.json()["message"]}')


def species(insert, order, details, lang):
    global config

    if insert is not None:
        r = http_post('/v1/species', {}, {
            'name': {
                lang: insert
            },
            'order_uuid': order
        })
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
    elif details is not None:
        r = http_get(f'/v1/species/{details}', {'X-Language': lang})
        if r is not None and r.status_code == 200:
            response = r.json()['result']
            print(f'id         {response["uuid"]}')
            print(f'name       {response["name"][lang]}')
            print(f'order      {response["order"]["name"][lang]}')
            wikipedia = r.json()['wikipedia']
            print(f'url        {wikipedia["url"]}')
            print(f'thumbnail  {wikipedia["thumbnail"]}')
            print(f'summary    {wikipedia["summary"]}')
            for img in wikipedia['images']:
                print(f'image      {img}')
            recordings = r.json()['recordings']
            if recordings:
                for rec in recordings:
                    print(f'recording  {rec["url"]}')
        else:
            response = r.json()
            err_print(f'err: {response["message"]}')
    else:
        r = http_get('/v1/species/list', {})
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
        else:
            data = r.json()
            t = []
            for species in data['result']:
                t.append([species['uuid'], species['name'][lang], species['order']['name'][lang]])
            print(tabulate.tabulate(t, headers=['id', f'name({lang})', f'order({lang})'], tablefmt='plain'))


def list(create, template, visibility, timespan, archive, rename, details, lang):
    global config
    if create is not None:
        v = 0
        if visibility == 'hidden':
            v = 1
        elif visibility == 'public':
            v = 2
        r = http_post('/v1/lists', {}, {
            'name': {
                lang: create
            },
            'template': template,
            'visibility': v
        })
        if r is not None:
            if r.status_code == 200:
                print(r.json()['uuid'])
            else:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif rename is not None:
        r = http_post(f'/v1/lists/{rename[0]}/rename', {}, {
            'name': {
                lang: rename[1]
            }
        })
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif archive:
        r = http_post(f'/v1/lists/{archive}/archive', {}, {})
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif timespan is not None:
        r = http_post(f'/v1/lists/{timespan[0]}/timespan', {}, {
            'from': timespan[1],
            'until': timespan[2]
        })
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif details:
        print("not yet implemented")
    else:
        r = http_get('/v1/lists/list', {})
        if r is not None:
            if r.status_code == 200:
                data = r.json()
                t = []
                for list in data['result']:
                    visibility = 'public'
                    if list['visibility'] == 0:
                        visibility = 'private'
                    elif list['visibility'] == 1:
                        visibility = 'hidden'
                    species = 0
                    if list['species'] is not None:
                        species = len(list['species'])
                    t.append([list['uuid'], list['name'][lang], visibility, species])
                print(tabulate.tabulate(t, headers=['id', f'name({lang})', 'visibility', 'species'], tablefmt='plain'))
            else:
                err_print(f'err: {r.json()["message"]}')


def report(create, species, age, method, gender, activity, quantity, position, comment, lists, _from, until, archive, details, tags, add_tag, remove_tag, lang, notification_comment, silent):
    if create:
        report = {
            'species_uuid': species,
            'age_uuid': str(age or ''),
            'method_uuid': str(method or ''),
            'gender_uuid': str(gender or ''),
            'activity_uuid': str(activity or ''),
            'comment': str(comment or ''),
            'list_uuids': lists,
            'from': str(_from or ''),
            'until': str(until or ''),
            'notification_comment': str(notification_comment or ''),
            'silent': silent,
            'tag_uuids': tags
        }
        if quantity is not None:
            report['quantity'] = quantity
        if position is not None:
            report['latitude'] = position[0]
            report['longitude'] = position[1]
        r = http_post('/v1/reports', {}, report)
        if r is not None:
            if r.status_code == 200:
                print(r.json()['uuid'])
            else:
                err_print(f'err: {r.json()["message"]}')
    elif archive:
        r = http_post(f'/v1/reports/{archive}/archive', {}, {})
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif details is not None:
        r = http_get(f'/v1/reports/{details}', {})
        if r is not None and r.status_code == 200:
            response = r.json()['result']
            position = ''
            method = ''
            activity = ''
            age = ''
            gender = ''
            lists = []
            tags = []
            if response['lists']:
                for l in response['lists']:
                    lists.append(l['name'][lang])
            if response['tags']:
                for t in response['tags']:
                    tags.append(t['name'])
            if response['latitude'] and response['longitude']:
                position = f'{response["latitude"]},{response["longitude"]}'
            if response['method']:
                method = response['method']['name'][lang]
            if response['activity']:
                activity = response['activity']['name'][lang]
            if response['age']:
                age = response['age']['name'][lang]
            if response['gender']:
                gender = response['gender']['name'][lang]
            print(f'id        {response["uuid"]}')
            print(f'name      {response["species"]["name"][lang]}')
            print(f'quantity  {response["quantity"]}')
            print(f'from      {get_date(response["from"])}')
            print(f'until     {get_date(response["until"])}')
            print(f'position  {position}')
            print(f'method    {method}')
            print(f'activity  {activity}')
            print(f'age       {age}')
            print(f'gender    {gender}')
            print(f'lists     {";".join(lists)}')
            print(f'tags      {";".join(tags)}')
            print(f'comment   {response["comment"]["text"]}')
        else:
            response = r.json()
            err_print(f'err: {response["message"]}')
    elif add_tag:
        r = http_post(f'/v1/reports/{add_tag[0]}/add_tag', {}, {
            'tag_uuid': add_tag[1]
        })
        if r is not None:
            if r.status_code != 204:
                response = r.json()
                err_print(f'err: {response["message"]}')
    elif remove_tag:
        r = http_post(f'/v1/reports/{remove_tag[0]}/remove_tag', {}, {
            'tag_uuid': remove_tag[1]
        })
        if r is not None:
            if r.status_code != 204:
                response = r.json()
                err_print(f'err: {response["message"]}')
    else:
        r = http_get('/v1/reports/list', {})
        if r is not None:
            if r.status_code == 200:
                data = r.json()
                #print(data)
                t = []
                for list in data['result']:
                    if list['gender'] is not None:
                        gender = list['gender']['name'][lang]
                    if list['age'] is not None:
                        age = list['age']['name'][lang]
                    if list['activity'] is not None:
                        activity = list['activity']['name'][lang]
                    if list['method'] is not None:
                        method = list['method']['name'][lang]
                    t.append([list['uuid'], list['quantity'], list['species']['name'][lang], get_date(list['from']), get_date(list['until'])])
                print(tabulate.tabulate(t, headers=['id', 'quantity', f'species({lang})', 'from', 'until'], tablefmt='plain'))
            else:
                err_print(f'err: {r.json()["message"]}')


def whoami():
    r = http_get('/v1/users/whoami', {})
    if r is not None:
        if r.status_code == 200:
            data = r.json()
            user = data['user']
            p = 'basic'
            if user['permissions'] & 2 == 2:
                p = 'admin'
            print(f'id           {user["uuid"]}')
            print(f'displayname  {user["display_name"]}')
            print(f'role         {p}')
            print(f'last login   {user["last_seen_at"]}')
            print(f'reports      {len(data["reports"])}')
        else:
            err_print(f'err: {r.json()["message"]}')


def user(details):
    if details:
        r = http_get(f'/v1/users/{details}', {})
        if r is not None:
            if r.status_code == 200:
                data = r.json()
                user = data['user']
                p = 'basic'
                if user['permissions'] & 2 == 2:
                    p = 'admin'
                print(f'id           {user["uuid"]}')
                print(f'displayname  {user["display_name"]}')
                print(f'role         {p}')
                print(f'reports      {len(data["reports"])}')
                print(f'last login   {user["last_seen_at"]}')
            else:
                err_print(f'err: {r.json()["message"]}')
    else:
        r = http_get('/v1/users', {})
        if r is not None:
            if r.status_code == 200:
                data = r.json()
                t = []
                for user in data:
                    p = 'basic'
                    if user['permissions'] & 2 == 2:
                        p = 'admin'
                    t.append([user['uuid'], user['display_name'], p])
                print(tabulate.tabulate(t, headers=['id', 'displayname', 'role'], tablefmt='plain'))
            else:
                err_print(f'err: {r.json()["message"]}')


def named_record(insert, details, endpoint, lang):
    if insert is not None:
        r = http_post(endpoint, {}, {
            'name': {
                'en': insert
            }
        })
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
    elif details is not None: 
        print(f'{endpoint}/{details}')
        r = http_get(f'{endpoint}/{details}', {})
        if r is not None and r.status_code == 200:
            response = r.json()['result']
            print(f'id    {response["uuid"]}')
            print(f'name  {response["name"][lang]}')
        else:
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
            for d in data['result']:
                t.append([d['uuid'], d['name'][lang]])
            print(tabulate.tabulate(t, headers=['id', 'name'], tablefmt='plain'))


def gender(insert, details, lang):
    named_record(insert, details, '/v1/genders', lang)


def order(insert, details, lang):
    named_record(insert, details, '/v1/orders', lang)


def age(insert, details, lang):
    named_record(insert, details, '/v1/ages', lang)


def activity(insert, details, lang):
    named_record(insert, details, '/v1/activities', lang)


def method(insert, details, lang):
    named_record(insert, details, '/v1/methods', lang)


def translation(update, en, la, sv):
    if update is not None:
        r = http_post('/v1/translations/update', {}, {
            'uuid': update,
            'en': en,
            'sv': sv,
            'la': la
        });
        if r is not None and r.status_code != 204:
            response = r.json()
            err_print(f'err: {response["message"]}')
    else:
        r = http_get('/v1/translations/list', {})
        if r is not None and r.status_code != 200:
            response = r.json()
            err_print(f'err: {response["message"]}')
        else:
            data = r.json()
            t = []
            for translation in data['translations']:
                t.append([translation['uuid'], translation['en'], translation['sv'], translation['la']])
            print(tabulate.tabulate(t, headers=['id', 'en', 'sv', 'la'], tablefmt='plain'))


def tag(create, details, color, rename, archive):
    if create:
        c = '#ffffff'
        if color:
            c = color
        r = http_post('/v1/tags', {}, {
            'name': create,
            'color': c
        })
        if r is not None:
            if r.status_code == 200:
                print(r.json()['uuid'])
            else:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
        pass
    elif details:
        r = http_get(f'/v1/tags/{details}', {})
        if r:
            response = r.json()['result']
            print(f'id     {response["uuid"]}')
            print(f'name   {response["name"]}')
            print(f'color  {response["color"]}')
        else:
            err_print('err: internal error')
    elif rename:
        r = http_post(f'/v1/tags/{rename[0]}/rename', {}, {
            'name': rename[1]
        })
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
    elif archive:
        r = http_post(f'/v1/tags/{archive}/archive', {}, {})
        if r is not None:
            if r.status_code != 204:
                err_print(f'err: {r.json()["message"]}')
        else:
            err_print('err: internal error')
        pass
    else:
        r = http_get('/v1/tags/list', {})
        response = r.json()
        t = []
        for tag in response['result']:
            t.append([tag['uuid'], tag['name']])
        print(tabulate.tabulate(t, headers=['id', 'name'], tablefmt='plain'))


def zone(insert, polygon, type):
    if insert:
        t = 4
        if type == 1:
            t = 'country'
        elif type == 'county':
            t = 2
        elif type == 'municipality':
            t = 3
        r = http_post(f'/v1/zones', {}, {
            'name': {
                'en': insert
            },
            'polygon': polygon,
            'type': t
        })
        if r is not None:
            if r.status_code == 200:
                print(r.json()['uuid'])
            else:
                err_print(f'err: {r.json()["message"]}')
    else:
        pass


def migrate(email, password):
    r = http_post('/v1/users/migrate', {}, {
        'old_username': email,
        'old_password': password
    })
    if not r:
        err_print(f'err: {r.json()["message"]}')
    else:
        print("yay")


def main():
    global config
    config.open(os.path.join(os.path.expanduser('~'), '.config/300sf/config.yml'))
    lang = config.get('language')
    if not lang:
        lang = 'en'

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

    api_key_parser = subparsers.add_parser('apikey', help='API key.')
    api_key_parser.add_argument('--create', dest='create', help='Create new API key.')
    api_key_parser.add_argument('--revoke', dest='revoke', help='Revoke API key.')
#    api_key_parser.add_argument('--scopes', dest='scopes', help='Scopes')

    user_parser = subparsers.add_parser('user', help='User.')
    user_parser.add_argument('--details', dest='details', help='Details.')

    template_parser = subparsers.add_parser('template', help='Template.')
    template_parser.add_argument('--create', dest='create', help='Create new template.')
    template_parser.add_argument('--description', dest='description', help='Create new template.')
    template_parser.add_argument('--visibility', dest='visibility', choices=['private', 'hidden', 'public'], help='Visibility.')
    template_parser.add_argument('--add', dest='add', nargs=2, help='Add species.')
    template_parser.add_argument('--remove', dest='remove', nargs=2, help='Remove species.')
    template_parser.add_argument('--archive', dest='archive', help='Archive.')
    template_parser.add_argument('--rename', dest='rename', nargs=2, help='Rename.')
    template_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')
    template_parser.add_argument('--details', dest='details', help='Details.')

    ## TODO: Add list_parser.
    list_parser = subparsers.add_parser('list', help='List.')
    list_parser.add_argument('--create', dest='create', help='Create new list.')
    list_parser.add_argument('--template', dest='template', help='Base list on this template.')
    list_parser.add_argument('--visibility', dest='visibility', choices=['private', 'hidden', 'public'], help='Visibility.')
    list_parser.add_argument('--timespan', dest='timespan', nargs=3, help='Timespan when the list can be reported to.')
    list_parser.add_argument('--archive', dest='archive', help='Archive.')
    list_parser.add_argument('--rename', dest='rename', nargs=2, help='Rename.')
    list_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')
    list_parser.add_argument('--details', dest='details', help='Details.')

    species_parser = subparsers.add_parser('species', help='Species.')
    species_parser.add_argument('--insert', dest='insert', help='Add species.')
    species_parser.add_argument('--order', dest='order', help='Order.')
    species_parser.add_argument('--details', dest='details', help='Details.')
    species_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')

    reports_parser = subparsers.add_parser('report', help='Reports.')
    reports_parser.add_argument('--create', dest='create', action='store_true', help='Create new report.')
    reports_parser.add_argument('--species', dest='species', help='Species.')
    reports_parser.add_argument('--age', dest='age', help='Age.')
    reports_parser.add_argument('--method', dest='method', help='Method.')
    reports_parser.add_argument('--gender', dest='gender', help='Gender.')
    reports_parser.add_argument('--activity', dest='activity', help='Activity.')
    reports_parser.add_argument('--quantity', dest='quantity', help='Quantity.')
    reports_parser.add_argument('--position', dest='position', nargs=2, help='Position WGS84.')
    reports_parser.add_argument('--comment', dest='comment', help='Comment.')
    reports_parser.add_argument('--lists', dest='lists', nargs='*', help='Comment.')
    reports_parser.add_argument('--from', dest='_from', help='From.')
    reports_parser.add_argument('--until', dest='until', help='Until.')
    reports_parser.add_argument('--archive', dest='archive', help='Archive.')
    reports_parser.add_argument('--details', dest='details', help='Details.')
    reports_parser.add_argument('--tags', dest='tags', nargs='*', default=[], help='Tags.')
    reports_parser.add_argument('--add-tag', dest='add_tag', nargs=2, help='Add tag.')
    reports_parser.add_argument('--remove-tag', dest='remove_tag', nargs=2, help='Remove tag.')
    reports_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')
    reports_parser.add_argument('--notification-comment', dest='notification_comment', help='Notification comment.')
    reports_parser.add_argument('--silent', dest='silent', action='store_true', default=False, help='Silent.')

    gender_parser = subparsers.add_parser('gender', help='Gender.')
    gender_parser.add_argument('--insert', dest='insert', help='Add gender.')
    gender_parser.add_argument('--details', dest='details', help='Details.')
    gender_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')

    age_parser = subparsers.add_parser('age', help='Gender.')
    age_parser.add_argument('--insert', dest='insert', help='Add age.')
    age_parser.add_argument('--details', dest='details', help='Details.')
    age_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')

    order_parser = subparsers.add_parser('order', help='Order.')
    order_parser.add_argument('--insert', dest='insert', help='Add order.')
    order_parser.add_argument('--details', dest='details', help='Details.')
    order_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')

    activity_parser = subparsers.add_parser('activity', help='Activity.')
    activity_parser.add_argument('--insert', dest='insert', help='Add activity.')
    activity_parser.add_argument('--details', dest='details', help='Details.')
    activity_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')

    method_parser = subparsers.add_parser('method', help='Method.')
    method_parser.add_argument('--insert', dest='insert', help='Add method.')
    method_parser.add_argument('--details', dest='details', help='Details.')
    method_parser.add_argument('--lang', dest='lang', default=lang, help='Language.')

    zone_parser = subparsers.add_parser('zone', help='Zone.')
    zone_parser.add_argument('--insert', dest='insert', help='Add zone.')
    zone_parser.add_argument('--polygon', dest='polygon', help='Polygon.')
    zone_parser.add_argument('--type', dest='type', choices=['country', 'county', 'municipality', 'habitat'], help='Type.')

    translation_parser = subparsers.add_parser('translation', help='Translation.')
    translation_parser.add_argument('--update', dest='update', help='Update translation uuid.')
    translation_parser.add_argument('--en', dest='en', help='English.')
    translation_parser.add_argument('--la', dest='la', help='Swedish.')
    translation_parser.add_argument('--sv', dest='sv', help='Latin.')

    tag_parser = subparsers.add_parser('tag', help='Tag.')
    tag_parser.add_argument('--create', dest='create', help='Create.')
    tag_parser.add_argument('--details', dest='details', help='Details.')
    tag_parser.add_argument('--color', dest='color', help='Color.')
    tag_parser.add_argument('--rename', dest='rename', nargs=2, help='Rename')
    tag_parser.add_argument('--archive', dest='archive', help='Archive.')

    subparsers.add_parser('whoami', help='Who am I.')

    data_parser = subparsers.add_parser('migrate', help='Migrate.')
    data_parser.add_argument('--email', dest='email', help='Email.')
    data_parser.add_argument('--password', dest='password', help='Password.')

    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    kwargs = vars(parser.parse_args())
    globals()[kwargs.pop('subparser')](**kwargs)


if __name__ == '__main__':
    main()

