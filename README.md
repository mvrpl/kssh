[![Release KSSH](https://github.com/mvrpl/kssh/actions/workflows/main.yaml/badge.svg)](https://github.com/mvrpl/kssh/actions/workflows/main.yaml)
[![Dependabot](https://badgen.net/badge/Dependabot/enabled/green?icon=dependabot)](https://dependabot.com/)
[![Scoop Version](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fraw.githubusercontent.com%2Fmvrpl%2Fwindows-apps%2Frefs%2Fheads%2Fmain%2Fbucket%2Fkssh.json&query=%24.version&style=flat&label=Scoop%20Version&link=https%3A%2F%2Fgithub.com%2Fmvrpl%2Fwindows-apps%2Fblob%2Fmain%2Fbucket%2Fkssh.json&color=%23012456)](https://github.com/mvrpl/windows-apps/blob/main/bucket/kssh.json)
[![Brew Version](https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fraw.githubusercontent.com%2Fmvrpl%2Funix-apps%2Frefs%2Fheads%2Fmain%2Fversions.json&query=%24.kssh&style=flat&label=Brew%20Version&color=%23701516&link=https%3A%2F%2Fgithub.com%2Fmvrpl%2Funix-apps%2Fblob%2Fmain%2FFormula%2Fkssh.rb)](https://github.com/mvrpl/unix-apps/blob/main/Formula/kssh.rb)


# SSH client with AWS KMS

## Installation
### With Go
```sh
go install github.com/mvrpl/kssh@latest
```
### MS Windows
```sh
scoop bucket add mvrpl https://github.com/mvrpl/windows-apps
scoop install mvrpl/kssh
```
### Unix
```sh
brew tap mvrpl/unix-apps https://github.com/mvrpl/unix-apps
brew install kssh
```
---
You can set AWS KMS resource ID or Alias:

    export KSSH_KEY=63c5fc45-f568-430e-89f5-3t92d7491f5e

Supported Cloud AWS KMS:

![AWS KMS Key](aws_kms_key.png)

## Support

    Linux, Mac OS and Microsoft Windows

## authorized_key

Print public key:

    kssh --authorized_key
    ecdsa-sha2-nistp256 AAAAzzz

You can copy the public key to ~/.ssh/authorized_keys in your home directory on the remote machine.

## ssh login

    kssh username@hostname

## usage

    kssh --help
