# 300sf CLI

Command line client for the 300sf API.

## Install

```bash
pip install ./cli        # from the repository root; installs the `300sf` command
```

Configure it in `~/.config/300sf/config.yml`:

```yaml
api-endpoint: https://300sf.se
client-id: <client id>
client-secret: <prefix>.<secret>
language: sv            # optional, en | sv | de
```

The session is stored in `~/.config/300sf/session.json` (mode 600).

## Tests

```bash
cd cli && python -m unittest discover -s tests
```

## Authentication
### Login
```bash
$ echo mysecretpassword > password
$ gpg2 --encrypt ./password
$ rm ./password
$ gpg2 --decrypt --quiet ./password.gpg | 300sf login --email me@example.com  
```

### Log out
```
$ 300sf logout
```

## Templates
### Create a Template
```bash
$ 300sf template --create "Svenska Fåglar" --visibility public
```

### List Templates
```bash
$ 300sf template
uuid                              name(en)        visibility      species
2409476e187f4acf9163510ebd200c58  Svenska Fåglar  public                0
```

### Rename a Template
```bash
$ 300sf template --rename 2409476e187f4acf9163510ebd200c58 "300 Svenska Fåglar"
```

### Add Species to Template
```bash
$ 300sf template --add 2409476e187f4acf9163510ebd200c58 325af33d039b42e58da2f194a6fcce20
```

### Remove Species from Template
```bash
$ 300sf template --remove 2409476e187f4acf9163510ebd200c58 325af33d039b42e58da2f194a6fcce20
```

### Archive a Template
```bash
$ 300sf template --archive 2409476e187f4acf9163510ebd200c58
```

## Lists
### Create an Empty List
```bash
$ 300sf list --create "My Epic List" --visibility hidden
```

### Create a List from a Template
```bash
$ 300sf list --create "My Epic list" --template 2409476e187f4acf9163510ebd200c58 --visibility hidden
```

### List Lists
```bash
$ 300sf list
```

### Rename a List
```bash
$ 300sf list --rename <list-uuid> "My List"
```

### Set List Start and End Date
```bash
$ 300sf list --timespan <list-uuid> "2023-01-01" "2023-12-31T23:59:00"
```

### Archive a List
```bash
$ 300sf list --archive <list-uuid>
```
