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
