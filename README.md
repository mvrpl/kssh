# SSH client with AWS KMS

    $ go install github.com/mvrpl/kssh@v1.0.0

You can set AWS KMS resource ID:

    $ export KSSH_KEY_ID=63c5fc45-f568-430e-89f5-3t92d7491f5e

Supported Cloud KMS algorithm:

- RSA_2048

## authorized_key

Print public key:

    $ kssh --authorized_key
    ecdsa-sha2-nistp256 AAAAzzz

You can copy the public key to ~/.ssh/authorized_keys in your home directory on the remote machine.

## ssh login

    $ kssh username@hostname

## usage

    $ kssh --help