# Installing Codapi

Make sure you install Codapi on a separate machine — this is a must for security reasons. Do not store any sensitive data or credentials on this machine. This way, even if someone runs malicious code that somehow escapes the isolated environment, they won't have access to your other machines and data.

Steps for Debian (11+) or Ubuntu (20.04+).

1. Install necessary packages (as root):

```sh
apt update && apt install -y ca-certificates curl docker.io make unzip
systemctl enable docker.service
systemctl restart docker.service
```

2. Create Codapi user (as root):

```sh
useradd --groups docker --shell /usr/bin/bash --create-home --home /opt/codapi codapi
```

3. Verify that Docker is working (as codapi):

```sh
docker run hello-world
```

4. Install Codapi (as codapi):

```sh
cd /opt/codapi
curl -L -o codapi.tar.gz "https://github.com/nalgeon/codapi/releases/download/v0.11.0/codapi_0.11.0_linux_amd64.tar.gz"
tar xvzf codapi.tar.gz
rm -f codapi.tar.gz
```

5. Build the sample `ash` sandbox image:

```sh
cd /opt/codapi
docker build --file sandboxes/ash/Dockerfile --tag codapi/ash:latest sandboxes/ash
```

6. Verify that Codapi starts without errors (as codapi):

```sh
cd /opt/codapi
./codapi
```

It should list `ash` in both `boxes` and `commands`:

```
2023/09/16 15:18:05 codapi 20230915:691d224
2023/09/16 15:18:05 listening on port 1313...
2023/09/16 15:18:05 workers: 8
2023/09/16 15:18:05 boxes: [alpine]
2023/09/16 15:18:05 commands: [sh]
```

Stop it with Ctrl+C.

7. Configure Codapi as systemd service (as root):

```sh
mv /opt/codapi/codapi.service /etc/systemd/system/
chown root:root /etc/systemd/system/codapi.service
systemctl enable codapi.service
systemctl start codapi.service
```

Verify that the Codapi service is running:

```sh
systemctl status codapi.service
```

Should print `active (running)`:

```
codapi.service - Code playgrounds
    Loaded: loaded (/etc/systemd/system/codapi.service; enabled; preset: enabled)
    Active: active (running)
...
```

8. Verify that Codapi is working:

```sh
curl -H "content-type: application/json" -d '{ "sandbox": "ash", "command": "run", "files": {"": "echo hello" }}' http://localhost:1313/v1/exec
```

Should print `ok` = `true`:

```json
{
    "id": "ash_run_dd27ed27",
    "ok": true,
    "duration": 650,
    "stdout": "hello\n",
    "stderr": ""
}
```
