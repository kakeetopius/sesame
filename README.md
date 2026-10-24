# sesame

A REST application (with a built-in client) that dynamically updates firewall rules on a Linux server so that only your IP can reach SSH.

Works with `firewalld` or `ufw`.

sesame started as a fork of [firewalld-rest](https://github.com/prashantgupta24/firewalld-rest) and has since changed a bit specifically `ufw` support and certificate-based request signing instead of JWT.

## Purpose

The idea is a server that does not allow SSH from anywhere by default, so there is nothing to brute-force. The only way in is to ask the sesame server to open SSH for your IP. The server checks that the request is signed with a key it trusts, and if so, adds a firewall rule that lets only that IP reach the SSH port. When you are done, you ask sesame to remove the rule and the server is closed off again.

You still need your SSH credentials to log in. sesame only decides who is allowed to try.

## Comparison with fail2ban

`fail2ban` is **reactive**: SSH is open to everyone, and addresses are banned after repeated failed logins. sesame is **proactive**: SSH is closed to everyone, and you open it for your own IP before you connect.

|             | `fail2ban`                               | `sesame`                                       |
| ----------- | ---------------------------------------- | ---------------------------------------------- |
| Approach    | Reactive: bans IPs after failed attempts | Proactive: allows only IPs you explicitly add  |
| SSH exposed | To everyone, until they get banned       | To nobody, except IPs you have allowed         |
| Rule change | Automatic, triggered by failures         | Manual, triggered by a signed request from you |

Both solve similar problem in different ways, and they can be used together.

## 1. Pre-requisites

1. A Linux server with **either** `firewalld` **or** `ufw` installed.
2. `root` access to the server. The sesame server runs the firewall commands, so it needs to be run as root.
3. [`step`](https://smallstep.com/docs/step-cli/installation/) to create keys and certificates (or `openssl`, see [3.2](#32-create-your-client-key-and-certificate)).
4. Go, to build the binary.

## 2. How it works

sesame has one binary with two subcommands:

```
sesame server    # runs on the Linux server you want to protect
sesame client    # runs on your own machine and sends signed requests to the server
```

### 2.1 Firewall backends

The server uses exactly one backend, chosen in the config (`firewalld` or `ufw`). One of them **must** be enabled, and they cannot both be enabled at once.

When you **allow** an IP, the server adds a rule for that IP on the SSH port (`ssh_port`, default `22`) and reloads the firewall. When you **deny** an IP, it removes the rule and reloads again.

| Backend     | Rule that is added                                                                                                                          |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `firewalld` | `firewall-cmd --permanent --zone=<zone> --add-rich-rule='rule family="ipv4" source address="<ip>/32" port protocol="tcp" port="22" accept'` |
| `ufw`       | `ufw allow from <ip> to any port 22 proto tcp`                                                                                              |

Both IPv4 and IPv6 are supported (`/32` and `/128` respectively).

**IP format:** pass a single address. A prefix is optional, but only `/32` (IPv4) or `/128` (IPv6) is accepted, so `203.0.113.7` and `203.0.113.7/32` are equivalent. Wider ranges such as `/24` are rejected, because the point is to open SSH for one machine only.

### 2.2 Database

The server keeps a small database (default `/var/sesame/sesame.db`) with every IP it has created a rule for. It is how the server keeps track of what it has opened. Keep it somewhere that survives reboots and cannot be deleted by other users.

An IP being in the database does not by itself mean it has SSH access. The database is only a record of the rules sesame has created.

### 2.3 Authentication

Every request must be signed with an **ed25519** key using [HTTP Message Signatures (RFC 9421)](https://www.rfc-editor.org/rfc/rfc9421.html). The server only accepts keys it has been told about.

- Each client creates its own ed25519 private key and certificate.
- The client **certificate** (never the private key) is copied to the server's `authorized_certs_dir`.
- When the server starts, it reads the certificates in that directory and stores each public key in a map indexed by its **SHA-256 hash**.
- Each request carries a signature named `magic-word` whose `keyid` is the hex-encoded SHA-256 hash of the client's public key. The server uses it to look up the public key and verify the signature.

What is signed:

| Request type | Signed components                                  |
| ------------ | -------------------------------------------------- |
| `GET`        | `@method`, `@authority`, `@path`                   |
| `POST`       | `@method`, `@authority`, `@path`, `content-digest` |

There is no shared secret on the server. To revoke a client, remove its certificate from `authorized_certs_dir`.

## 3. Installation

### 3.1 Build

```
go build
```

This produces a `sesame` binary. Copy it to the server and put it somewhere on the `PATH`:

```
scp sesame root@<server>:/usr/local/bin/sesame
```

You also need the binary on your own machine to use the client.

### 3.2 Create your client key and certificate

Do this on **your own machine**. First [install `step`](https://smallstep.com/docs/step-cli/installation/), then run:

```
step certificate create pius client.cert client.key \
  --profile self-signed --subtle \
  --kty OKP --curve Ed25519 \
  --no-password --insecure
```

What the command does:

| Part                         | Meaning                                                                                   |
| ---------------------------- | ----------------------------------------------------------------------------------------- |
| `pius`                       | The certificate's subject name. Use anything that identifies you or the machine.          |
| `client.cert` / `client.key` | Output files: the certificate (shared with the server) and the private key (kept secret). |
| `--profile self-signed`      | Creates a self-signed certificate if there is no CA to set up.                            |
| `--subtle`                   | Required by `step` when creating a self-signed certificate; it acknowledges the profile.  |
| `--kty OKP --curve Ed25519`  | Makes an **ed25519** key. sesame only accepts ed25519.                                    |
| `--no-password --insecure`   | Leaves the private key unencrypted, so the client can use it without a prompt.            |

Then lock the key down:

```
chmod 600 client.key
```

**Alternative with OpenSSL:**

```
openssl genpkey -algorithm ed25519 -out client.key
openssl req -new -x509 -key client.key -out client.cert -days 365 -subj "/CN=pius"
```

- `genpkey -algorithm ed25519` generates the private key.
- `req -new -x509` creates a self-signed certificate from that key, valid for 365 days, with `pius` as its common name.

### 3.3 Create the server key and certificate

If you run the server over TLS (recommended whenever it is reachable from the internet), the server needs its own key and certificate. Create them on the server:

```
sudo mkdir -p /etc/sesame/authorized_certs /var/sesame

sudo step certificate create sesame-server /etc/sesame/sesame.cert /etc/sesame/sesame.key \
  --profile self-signed --subtle \
  --san <server-ip-or-hostname> \
  --no-password --insecure

sudo chmod 600 /etc/sesame/sesame.key
```

- `--san` adds the IP or hostname clients will connect to. Clients verify that the certificate matches the address they connect to.
- The first command also creates the directories sesame uses by default (`/etc/sesame/authorized_certs` for client certs and `/var/sesame` for the database).
- The server passes these files straight to Go's http server, so any key type that Go supports works here. Only the **client** keys must be ed25519.

If your clients do not trust this self-signed certificate, they can pass `--skip-verify` (see [Security notes](#6-security-notes)).

### 3.4 Authorize your client on the server

Copy your client **certificate** (not the key) into the server's authorized certs directory:

```
scp client.cert root@<server>:/etc/sesame/authorized_certs/pius.cert
```

Use a different file name for each client.

If you ever need to sign requests yourself (see [5.2](#52-signing-requests-yourself)), you can compute your `keyid` from the certificate:

```
openssl x509 -in client.cert -pubkey -noout \
  | openssl pkey -pubin -outform DER \
  | tail -c 32 \
  | sha256sum | awk '{print $1}'
```

How this works:

1. `openssl x509 ... -pubkey -noout` extracts the public key from the certificate in PEM form.
2. `openssl pkey -pubin -outform DER` converts it to DER. For ed25519 this is a 44-byte structure whose last 32 bytes are the raw public key.
3. `tail -c 32` keeps just those 32 raw key bytes.
4. `sha256sum | awk '{print $1}'` hashes them and prints the hex digest, which is the `keyid`.

### 3.5 Configure the server

sesame uses [Viper](https://github.com/spf13/viper), so the config file can be any format Viper supports (TOML, YAML, JSON, ...). Pass it with `--config`. The examples here use TOML.

`/etc/sesame/config.toml`:

```toml
address = "0.0.0.0"
port = 8080
ssh_port = 22
insecure = false                 # false = serve over TLS
db = "/var/sesame/sesame.db"
cert_file = "/etc/sesame/sesame.cert"
key_file = "/etc/sesame/sesame.key"
authorized_certs_dir = "/etc/sesame/authorized_certs"

[firewalld]
enabled = true
zone = "public"

[ufw]
enabled = false
```

Use **either** `[firewalld]` **or** `[ufw]`: exactly one must be enabled.

**Defaults** (used when a value is not set):

| Key                    | Default                              |
| ---------------------- | ------------------------------------ |
| `address`              | `0.0.0.0`                            |
| `port`                 | `8080`                               |
| `ssh_port`             | `22`                                 |
| `db`                   | `/var/sesame/sesame.db`              |
| `authorized_certs_dir` | `/etc/sesame/authorized_certs`       |
| `key_file`             | `/etc/sesame/sesame.key`             |
| `cert_file`            | `/etc/sesame/sesame.cert`            |
| `firewalld.zone`       | `public` (when firewalld is enabled) |

`insecure = true` serves plain HTTP instead of TLS. In that case `cert_file` and `key_file` are not needed. Only use it for local testing or behind something that terminates TLS for you.

**Server flags** (these can override the config file):

| Flag           | Meaning                                  |
| -------------- | ---------------------------------------- |
| `--config`     | Path to the server config file           |
| `-a`, `--addr` | Address to listen on                     |
| `-p`, `--port` | Port to listen on                        |
| `--ssh-port`   | Port the SSH daemon is listening on (22) |

### 3.6 Run the server with systemd

Create `/etc/systemd/system/sesame.service`:

```ini
[Unit]
Description=sesame server
After=network-online.target firewalld.service
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/sesame server --config /etc/sesame/config.toml
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

- `User=root` is required because the server runs firewall commands.
- `Restart=on-failure` brings the server back up if it crashes, which matters because it is your only way in.

Start it and enable it at boot:

```
sudo systemctl daemon-reload
sudo systemctl enable --now sesame
```

Check the logs with:

```
journalctl -u sesame -f
```

### 3.7 Set up the firewall

Note that the steps below close SSH to everyone, so the order matters. **Keep your current SSH session open until you have tested an `allow` from the client.** Already-established connections keep working when you remove the SSH rule, but a new one will be refused.

1. **Open the sesame port** so you can reach the server once SSH is closed.
2. **Remove SSH** from the firewall.
3. **Test** that `sesame client allow` brings SSH back for your IP.

**firewalld**

```
# 1. let anyone reach the sesame server (the port from your config)
firewall-cmd --zone=public --add-port=8080/tcp --permanent

# 2. remove SSH access for everyone
firewall-cmd --zone=public --remove-service=ssh --permanent

firewall-cmd --reload

# confirm: ssh should no longer be listed
firewall-cmd --zone=public --list-all
```

**ufw**

```
# 1. let anyone reach the sesame server
sudo ufw allow 8080/tcp

# 2. remove SSH access for everyone
sudo ufw delete allow ssh        # or: sudo ufw delete allow 22/tcp

sudo ufw enable

# confirm
sudo ufw status
```

If your SSH daemon listens on a port other than 22, remove that port instead and set `ssh_port` in the sesame config.

Then, from a **new terminal**, try SSH to the server. It should be refused. Next, run the client (see below) and try again. It should work.

## 4. Using the client

Create a client config on your machine, for example `client.toml`:

```toml
address = "203.0.113.10"      # the server's IP or hostname
port = 8080
cert_file = "certs/pius.cert"
private_key_file = "certs/pius.key"
```

All four values are required.

**Client flags** (these override the config file):

| Flag            | Meaning                                                         |
| --------------- | --------------------------------------------------------------- |
| `--config`      | Path to the client config file                                  |
| `-a`, `--addr`  | Server address to connect to                                    |
| `-p`, `--port`  | Server port to connect to                                       |
| `--cert-file`   | Path to your client certificate                                 |
| `--key-file`    | Path to your client private key                                 |
| `--skip-verify` | Do not verify the server's certificate when connecting over TLS |

**Allow your IP**

First find your public IP (for example with `curl -4 ifconfig.me`), then:

```
sesame client --config client.toml allow 198.51.100.25
```

You can now SSH to the server from that IP.

**Deny your IP** when you are done:

```
sesame client --config client.toml deny 198.51.100.25
```

Everything can also be given on the command line without a config file:

```
sesame client -a 203.0.113.10 -p 8080 \
  --cert-file certs/pius.cert --key-file certs/pius.key \
  allow 198.51.100.25
```

## 5. REST API

You do not have to use `sesame client`. Any program that can send correctly signed requests will work, but your private key and certificate are always required.

### 5.1 Routes

| Purpose                            | Method   | Path          | Body                    | Client command        |
| ---------------------------------- | -------- | ------------- | ----------------------- | --------------------- |
| Index page                         | `GET`    | `/m1`         |                         | not yet               |
| Show all IPs present               | `GET`    | `/m1/ip`      |                         | not yet               |
| Add new IP (allow SSH)             | `POST`   | `/m1/ip`      | `{"ip": "203.0.113.7"}` | `sesame client allow` |
| Show if a particular IP is present | `GET`    | `/m1/ip/{ip}` |                         | not yet               |
| Delete IP (deny SSH)               | `DELETE` | `/m1/ip/{ip}` |                         | `sesame client deny`  |

### 5.2 Signing requests yourself

Requests follow RFC 9421. Sign these components with your ed25519 private key, under the signature name `magic-word`:

- `GET` and `DELETE`: `@method`, `@authority`, `@path`
- `POST`: the same, plus `content-digest` (a `Content-Digest` header computed over the body)

Set `keyid` to your key's hash (see [3.4](#34-authorize-your-client-on-the-server)). The result is two headers on the request:

```
Signature-Input: magic-word=("@method" "@authority" "@path");created=1767225600;keyid="<hex sha256 of your public key>"
Signature: magic-word=:<base64 signature>:
```

For a `POST`, the covered components also include `"content-digest"` and the request carries a `Content-Digest` header.

A few things to know:

- `created` must be recent. The server rejects signatures older than a short window (10 seconds by default), so your clock needs to be roughly in sync with the server's.
- `curl` cannot compute these signatures on its own. Use `sesame client`, or an RFC 9421 library such as [yaronf/httpsign](https://github.com/yaronf/httpsign) for Go, which sesame itself uses.

## 6. Security notes

- **Replay window:** a signed request is only accepted for a short time after it was created (10 seconds by default). This limits replay, but a captured request could still be replayed inside that window. Serving over TLS (`insecure = false`) keeps requests from being captured on the network in the first place. See [Future work](#7-future-work).
- **`--skip-verify`:** turns off verification of the server's TLS certificate in the client. Use it only for testing, or when you cannot trust the certificate any other way, because it removes protection against someone impersonating your server.
- **`insecure = true`:** serves plain HTTP. Do not use it on a server reachable from the internet.
- **Private keys:** keep `client.key` and `sesame.key` readable by you (or root) only (`chmod 600`).
- **Revoking access:** delete a client's certificate from `authorized_certs_dir`.

## 7. Future work

- Maybe more robust replay protection, such as server-side tracking of nonces so each signed request can only be used once.

## 8. Helpful links

- firewalld
  - Documentation: https://firewalld.org/documentation/
  - Tutorial: https://www.digitalocean.com/community/tutorials/how-to-set-up-a-firewall-using-firewalld-on-centos-7

  Useful commands:

  ```
  firewall-cmd --get-default-zone
  firewall-cmd --get-active-zones
  firewall-cmd --zone=public --list-all
  firewall-cmd --permanent --zone=public --list-rich-rules
  firewall-cmd --reload
  ```

- ufw
  - Guide: https://help.ubuntu.com/community/UFW

  Useful commands:

  ```
  sudo ufw status numbered
  sudo ufw allow from 203.0.113.7 to any port 22 proto tcp
  sudo ufw delete <rule-number>
  ```

## Credits

Based on [firewalld-rest](https://github.com/prashantgupta24/firewalld-rest) by prashantgupta24.
