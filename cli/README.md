# 300sf CLI

## Authentication
### Login
```bash
$ echo mysecretpassword > passsword
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
???
```

### Rename a List
```bash
$ 300sf list --rename ??? "My List"
```

### Set List Start and End Date
```bash
$ 300sf list --timespan ??? "2023-01-01" "2023-12-31T23:59:00"
```

### Archive a List
```bash
$ 300sf list --archive ???
```
