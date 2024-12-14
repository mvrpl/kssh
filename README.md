![build workflow](https://github.com/github/docs/actions/workflows/main.yml/badge.svg)
[![Dependabot](https://badgen.net/badge/Dependabot/enabled/green?icon=dependabot)](https://dependabot.com/)
![Scoop Version](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fraw.githubusercontent.com%2Fmvrpl%2Fwindows-apps%2Frefs%2Fheads%2Fmain%2Fkssh.json&query=%24.version&style=flat&label=Scoop%20Version&link=https%3A%2F%2Fgithub.com%2Fmvrpl%2Fwindows-apps%2Fblob%2Fmain%2Fkssh.json)


# SSH client with AWS KMS

    $ go install github.com/mvrpl/kssh@latest

You can set AWS KMS resource ID:

    $ export KSSH_KEY_ID=63c5fc45-f568-430e-89f5-3t92d7491f5e

Supported Cloud AWS KMS:

![AWS KMS Key](aws_kms_key.png)

## Support

    Linux, Mac OS and Microsoft Windows

## authorized_key

Print public key:

    $ kssh --authorized_key
    ecdsa-sha2-nistp256 AAAAzzz

You can copy the public key to ~/.ssh/authorized_keys in your home directory on the remote machine.

## ssh login

    $ kssh username@hostname

## usage

    $ kssh --help
