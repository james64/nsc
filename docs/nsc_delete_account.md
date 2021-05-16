## nsc delete account

Delete an account and associated users

### Synopsis

Delete an account and associated users

```
nsc delete account [flags]
```

### Examples

```
nsc delete account -n name
nsc delete account -i

```

### Options

```
  -F, --force         managed accounts must supply --force
  -h, --help          help for account
  -n, --name string   name of account to delete
  -R, --revoke        revoke users before deleting (default true)
  -C, --rm-creds      delete users creds
  -D, --rm-nkey       delete user keys
```

### Options inherited from parent commands

```
  -i, --interactive          ask questions for various settings
  -K, --private-key string   private key
```

### SEE ALSO

* [nsc delete](nsc_delete.md)	 - Delete imports and exports

###### Auto generated by spf13/cobra on 18-Mar-2021