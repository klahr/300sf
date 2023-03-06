# 300sf CLI

```
$ echo mysecretpassword > passsword
$ gpg2 --encrypt ./password
$ rm ./password
$ gpg2 --decrypt --quiet ./password.gpg | 300sf login --email me@example.com  
```
