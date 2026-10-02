# mws-api

Go port of the Lucee MWS-API project. This port should provide better performance on a newer language with better support.
Currently only handles the `/schedule` endpoint, but will be updated to include `/catalog` shortly after initial release.

## Installing on RHEL 10

### Requirements
The only real dependency for the service is the `age` encryption package. Installing that package requires enabling
both the code-ready-builder and EPEL (Extra Packages for Enterprise Linux) repositories. Do so via:
```bash
sudo subscription-manager repos --enable codeready-builder-for-rhel-10-$(arch)-rpms
sudo dnf install https://dl.fedoraproject.org/pub/epel/epel-release-latest-10.noarch.rpm
```

Afterwards install `age`
```bash
sudo dnf install age
```

### Service Setup

#### Configuration

Create a service account that will own the age file and start the API service.
```bash
sudo useradd --system --no-create-home --shell /sbin/nologin mwsapi
# verify with
getent passwd mwsapi
```

Create the service configuration directory.
```bash
sudo mkdir -p /etc/mwsapi/age
sudo chown root:root /etc/mwsapi
sudo chmod 755 /etc/mwsapi
```

Restrict the age directory.
```bash
sudo chown mwsapi:mwsapi /etc/mwsapi/age
sudo chmod 700 /etc/mwsapi/age
```

Generate the age key as the mwsapi user.
```bash
sudo -u mwsapi age-keygen -o /etc/mwsapi/age/identity
```

Fix permissions on the key.
```bash
sudo chmod 600 /etc/mwsapi/age/identity
sudo chown mwsapi:mwsapi /etc/mwsapi/age/identity
```

Note the recipient, as it will be used later.
```bash
sudo age-keygen -y /etc/mwsapi/age/identity
```

Encrypt the environment variables file for the recipient key.
```bash
age -r $(sudo age-keygen -y /etc/mwsapi/age/identity) -o .env.age .env
sudo install -o root -g mwsapi -m 640 .env.age /etc/mwsapi/.env.age
# you can test with this:
sudo -u mwsapi age -d -i /etc/mwsapi/age/identity /etc/mwsapi/.env.age
```

#### Binary/Service

Install the binary from GitHub.
```bash
curl -L https://github.com/cself-sdccd-edu/mws-api/releases/latest/download/mws-api -o /tmp/mws-api
sudo install -m 755 /tmp/mws-api /usr/local/bin/mws-api
```

Create a command wrapper script: `vim ./mws-api-start.sh`
```bash
#!/bin/bash
set -euo pipefail

set -a
source <(age -d -i /etc/mwsapi/age/identity /etc/mwsapi/.env.age)
set +a

exec /usr/local/bin/mws-api
```

Install the script.
```bash
sudo install -m 755 -o root -g root mws-api-start.sh /usr/local/bin/mws-api-start
```

Create the systemd unit file: `sudo vim /etc/systemd/system/mws-api.service`
```bash
[Unit]
Description=MWS API
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
User=mwsapi
Group=mwsapi
WorkingDirectory=/etc/mwsapi
ExecStart=/usr/local/bin/mws-api-start
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Copy existing app.json configuration file to the server and install it.
```bash
sudo install -o mwsapi -g mwsapi -m 640 app.json /etc/mwsapi/config.json
```

Reload systemd. Start the service. Verify it is running as expected.
```bash
sudo systemctl daemon-reload
sudo systemctl start mws-api.service
systemctl status mws-api.service
```

Allow external access with firewall-cmd.
```bash
sudo firewall-cmd --permanent --add-port=6767/tcp
# reload and verify
sudo firewall-cmd --reload
sudo firewall-cmd --list-all
```

Verify access from another machine.
```bash
nc -vz mwsapi01.sdccd.edu 6767
```




