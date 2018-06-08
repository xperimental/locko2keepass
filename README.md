# locko2keepass

This is a converter utility for converting exports from [Locko](https://binarynights.com/blog/locko-super-slick-password-manager-and-filevault/) which is defunct now to a Keepass 2 compatible file.

## Usage

```bash
locko2keepass passwords.lckexp
# This will create a passwords.lckexp.kdbx file
```

The password of the generated file will be `default`. Open the database in Keepass or a compatible program and change the password.

