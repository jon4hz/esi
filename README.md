# esi (env-secret-injector)

## ☕ About
ESI is a cli tool that allows you to fetch secrets from password managers and inject them to other processes in form of environment variables, 
config files or from stdin.  
This unlocks a secure development and debug process, where you dont have to store any credentials locally but only fetch them on demand.


## 🚀 Installation
```
# using go directly
$ go install github.com/jon4hz/esi@latest

# local pkg manager
$ export VERSION=v1.4.2

## debian / ubuntu
$ dpkg -i esi-$VERSION-linux-amd64.deb

## rhel / sles
$ rpm -i esi-$VERSION-linux-amd64.rpm

## alpine
$ apk add --allow-untrusted esi-$VERSION-linux-amd64.apk
```
All releases can be found [here](https://github.com/jon4hz/esi/releases)


## ✨ Usage
### 🗿 Command mode
By default, `esi` will take your input as a single command, without any support for shell specific features like aliases, pipes or redirects.
```bash
$ esi -- echo You cannot pipe my output \:\(
```

### 🐚 Shell mode
If you use `esi`'s shell mode, `esi` will spawn your command in a subshell and support all the fancy stuff your heart might desire.
```bash
$ esi shell -- "env | grep MY_SECRET || echo could not find my secret."
```

> **NOTE:** Make sure to put your command in quotes and escape where escape is needed!



## 📝 Config
ESI searches for a config file in the following locations:

1. `./esi.yml`
2. `~/.config/esi/esi.yml`
3. `/etc/esi/esi.yml`


First come first serve, if you dont like that, use the `--config` flag to specify an exact location.


### Secret Server Config
The secret server config contains all information `esi` needs in order to connect to the TSS.

| Name | Description | Value
|-|-|-|
| `url` | URL to the secret server | `https://my-secret-server.com`
| `ttl` | expiration time of your access token (seconds) | `7200`


### Secrets config
In order to inject any secrets, you need to tell `esi` which ones it should fetch.

| Name | Description | Value
|-|-|-|
|`id`| A **unique** id of the secret. <br> You will reference the secret by this id in the injector config | `""`
|`secret_id` | Secret ID from TSS (in url of secret) | `0`
|`field` | Field from the secret that contains the desired value | `""`

> **NOTE:** esi will only fetch secrets that are actually used by injectors.


### Injector config
An injector defines how the secrets are passed to your application. An injector accepts multiple configs, allowing you to inject an arbitrary number of secrets.


#### Injector group
To keep things organized, you can group your injectors. When executing `esi` without any special settings, `esi` will interactively ask you which injector to use.

| Name | Description | Value
|-|-|-|
|`name`| Unique name of the group | `""`
|`selected` | Is this group selected by default? | `false`
|`injectors` | An array of injector configs | `[]`


#### Injectors

##### Env injector

| Name | Description | Value
|-|-|-|
`env_key`| name of the environment variable to be injected | `""`
`env_secret`| ID of the secret to be injected | `""`

##### Stdout injector

| Name | Description | Value
|-|-|-|
`stdout`| Print the secret to stdout | `false`
`stdout_secret` | ID of the secret to be injected | `""`

##### Config injector

The config injector is probably one of `esi`'s most advanced features.  
It allows you to create a config template which is deployed before the application starts and automatically cleaned up after.  

The path to the temporary config file will be stored in an environment variable.

The config is templated using go template (the same template engine used by helm e.g.).



| Name | Description | Value
|-|-|-|
`tmp_file` | Use the config injector | `false`
`tmp_file_secrets` | IDs of the secrets that you need for your config | `[]`
`tmp_file_tmpl` | The template of the temporary config file | `""`
`tmp_file_var` | Env var that contains the path to the config | `""`
`tmp_file_suffix` | Suffix of the temporary config file | `""`


> **NOTE:** 
To reference a secret by it's id, you can use the following pattern:
```
{{- with (index .Secrets "my-secret-id") -}}
{{ .Value }}
{{- end -}}
```


## 🔐 Authentication
First of all `esi` will ask you for a "local encryption password". This password will encrypt the TSS API token. You will have to enter this encryption password every 15 minutes, so choose something secure and memorable.

To authenticate against the TSS, you need to fetch an API token.  
1. Login to the TSS
2. User Preferences (click on your avatar)
3. "Generate API Token and Copy to Clipboard"

<details>
<summary>Obligatory XKCD</summary>
![Password Strength](assets/password_strength.png){width=75%}
</details>


## 🥁 Examples

### Config for Ansible Galaxy & Ansible Vault
Where's that pesky ansible vault password again? Ah yes - it's stored centrally on the secret server!

```yaml
---
secret_server:
  # url to the secret server
  url: https://my-secret-server.com
  # expiration time of your access token
  ttl: 7200  # 2h

# available secrets
secrets:
  - id: lxp/prod/aap-infra
    secret_id: 41611
    field: password

  - id: ansible-hub-token
    secret_id: 43751
    field: password


# create groups for multiple injectors
groups:
  - name: ansible-vaults
    selected: true
    injectors:
      # inject an ansible vault password:
      # usage: esi -- ansible-vault encrypt
      - name: lxp/prod/aap-infra
        configs:
          - tmp_file: true
            tmp_file_var: ANSIBLE_VAULT_PASSWORD_FILE
            tmp_file_secrets:
              - lxp/prod/aap-infra
            tmp_file_tmpl: |
              {{- with (index .Secrets "lxp/prod/aap-infra") -}}
              {{ .Value }}
              {{- end }}

  - name: ansible-cfgs
    injectors:
      # inject an ansible.cfg with preconfigure galaxy settings:
      # usage: esi -- ansible-galaxy install -r requirements.yml
      - name: simple-galaxy
        selected: true
        configs:
          - tmp_file: true
            tmp_file_var: ANSIBLE_CONFIG
            tmp_file_suffix: .cfg
            tmp_file_secrets:
              - ansible-hub-token
            tmp_file_tmpl: |
              [defaults]
              timeout = 30

              [galaxy]
              server_list = published,community,rh-certified,validated

              [galaxy_server.community]
              url=https://my.ansible-hub.com/api/galaxy/content/community/
              token={{- with (index .Secrets "ansible-hub-token") -}}{{ .Value }}{{- end }}

              [galaxy_server.published]
              url=https://my.ansible-hub.com/api/galaxy/
              token={{- with (index .Secrets "ansible-hub-token") -}}{{ .Value }}{{- end }}


              [galaxy_server.rh-certified]
              url=https://my.ansible-hub.com/api/galaxy/content/rh-certified/
              token={{- with (index .Secrets "ansible-hub-token") -}}{{ .Value }}{{- end }}

              [galaxy_server.validated]
              url=https://my.ansible-hub.com/api/galaxy/content/validated/
              token={{- with (index .Secrets "ansible-hub-token") -}}{{ .Value }}{{- end }}
```

### Add a Key to an SSH Agent
The sky is the limit when it comes to fancy stuff you can do with `esi`. This includes adding keys to an ssh-agent, even if they are passphrase protected.  
All you need is the correct `esi.yml` config and a small helper script. Sounds neat, doesn't it?

#### esi.yml
```yaml
---
secret_server:
  url: https://my-secret-server.com
  ttl: 7200 # 2h

secrets:
  - id: ssh-key-passphrase
    secret_id: 44047
    field: private-key-passphrase

  - id: ssh-key
    secret_id: 44047
    field: private-key

groups:
  - name: ssh
    injectors:
      - name: key
        configs:
          - stdout: true
            stdout_secret: ssh-key
      - name: passphrase
        configs:
          - stdout: true
            stdout_secret: ssh-key-passphrase
```

#### ./add-ssh-key.sh
```bash
#!/bin/bash
esi login
echo -e '#!/bin/bash\nesi --injector=ssh.passphrase -- echo' > /tmp/ssh_helper
chmod 700 /tmp/ssh_helper
esi --injector=ssh.key -- echo | DISPLAY=None SSH_ASKPASS="/tmp/ssh_helper" ssh-add -
rm /tmp/ssh_helper
```
