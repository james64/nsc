## nsc import user

Imports an user from a jwt or user and nkey from a creds file

### Synopsis

Imports an user from a jwt or user and nkey from a creds file

```
nsc import user --file <user-jwt/user-creds> [flags]
```

### Examples

```
nsc import user --file <account-jwt>
```

### Options

```
  -f, --file string   user jwt or creds to import
  -h, --help          help for user
      --overwrite     overwrite existing user
      --skip          skip validation issues
```

### Options inherited from parent commands

```
  -i, --interactive          ask questions for various settings
  -K, --private-key string   private key
```

### SEE ALSO

* [nsc import](nsc_import.md)	 - Import assets such as nkeys

###### Auto generated by spf13/cobra on 18-Mar-2021