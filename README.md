
[![MIT License](https://img.shields.io/apm/l/atomic-design-ui.svg?)](https://github.com/tterb/atomic-design-ui/blob/master/LICENSEs)
![Docker Image Size (tag)](https://img.shields.io/docker/image-size/psihoza/ddns-aws/latest)


```
 ______  ______  __   _ _______     _______ _  _  _ _______
 |     \ |     \ | \  | |______ ___ |_____| |  |  | |______
 |_____/ |_____/ |  \_| ______|     |     | |__|__| ______|
                                                           
```

# DDNS-AWS

DDNS-AWSD is a self-hosted Dynamic DNS server that uses Amazon AWS Route53.
## Features

- can be used as CLI tool, System service or Docker Image
- minimal (less than 10MB), no-dependency binary written in Go. Docker image also less than 10MB
- basic compatibility with Dynamic DNS Update API (https://help.dyn.com/remote-access-api/)
- CLI to list/add/remove managed A records from AWS Route 53
- CLI to manually update IP of selected A record


## Requirements
DDNS-AWS uses AWS Route53 DNS provider so you need at least 1 hosted zone on route 53 (which is free with AWS Free tier).

Next, you need AWS IAM user and policy configured, with minimal access rights for security reasons.
You don't want your whole AWS account compromised if your config file gets exposed somehow.

Configuring AWS is out of the scope of this document, but short howto is:

1) Login to AWS Console and open the service 'IAM Managemenent Console'
2) Click 'Policies' -> 'Create Policy' -> tab 'JSON'
3) copy paste this (tighten `Resource` more if you don't want to give access to all DNS resources):
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "elasticloadbalancing:DescribeLoadBalancers",
                "route53:ListHostedZones",
                "route53:ChangeResourceRecordSets",
                "route53:ListResourceRecordSets"
            ],
            "Resource": "*"
        }
    ]
}
```
4) In next steps, name and describe your policy and click to create it. 
5) Under `Users` click `Add Users` and fill in details (for credentials type choose 'programmatic access')
6) Next step is 'Permissions'. Click top box named 'attach  existing policy directly'
7) In the list below, search for the policy you just created and mark the checkbox next to it.
8) Finish creating user. Last steps are selfe-xplanatory.
9) You will be presented with access and secret keys. Copy them to safe place or download .csv

Now you have credentials needed for DDNS-AWS to connect to Route53 safely.

TLS Certificates are optional but highly recommended for production. This enables HTTPS and encryption.
You can create self signed certificate (https://www.linode.com/docs/guides/create-a-self-signed-tls-certificate/),
but idealy you should use free services like LetsEncrypt or similar (https://geekflare.com/free-ssl-tls-certificate/).
Process of renewing certificates can also be automated.
## Installation

There are 3 methods of installation and usage: CLI program, System Service or Docker image.

### Install as a CLI program

```bash
git clone https://github.com/psiho/ddns-aws
cd ddns-aws
make build
```

Now, edit `.ddns-aws.yaml` configuration file and then install with:
```bash
sudo make install
```

Edit config file `/etc/.ddns-aws.yaml` and enter your server details. Also enter your AWS credentials to be able to connect to Route53 (described in 'Requirements' section).

Alternatively, you can use environment variables to specify most of the settings. Those values will take precendence over ones in config file.
Also, note that DDNS-AWS will overwrite config file with values from ENV variables.

Environment variable names are:
```
DDNS_AWS_SERVER_PORT
SERVER_USERNAME
SERVER_PASSWORD
SERVER_CERT
SERVER_PRIVATEKEY
AWS_REGION
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
```
### Install as a service
This is a recommended method of installation if you plan to run it 24/7, which is expected way of using any DDNS service.

It should work on any Systemd linux variant.

After following above steps for installation as CLI, just add one more command:
```bash
sudo make service
```
start service with:
```bash
sudo systemnctl start ddns-aws
```
... then check status with:
```bash
sudo systemnctl status ddns-aws
```
If you want to enable automatic start at system startup:
```bash
sudo systemnctl enable ddns-aws
```
### Docker install
Pre-built Docker image and instructions are at https://hub.docker.com/r/psihoza/ddns-aws

If you want to build your own image, `Dockerfile` is available in the root directory. You can create docker image with:
```bash
make docker
```

You can run Docker image like:
```bash
docker run -it \
    -e AWS_REGION=<enter-aws-region> \
    -e AWS_ACCESS_KEY_ID=<your-key-from-aws-iam> \
    -e AWS_SECRET_ACCESS_KEY=<your-secret-key-from-aws-iam> \
    -e DDNS_AWS_SERVER_PORT=8125 \
    -e SERVER_USERNAME=<username-for-ddns-client> \
    -e SERVER_PASSWORD=<password-for-ddns-client> \
    -e SERVER_CERT=/certificates/<your-server-certificate> \
    -e SERVER_PRIVATEKEY=/certificates/<your-server-private-key> \
    --name=ddns-aws \
    -p 8125:8125 \
    -v <host-path-to-server-certificates>:/certificates \
    -d psihoza/ddns-aws
```

How to get AWS keys and TLS Certificates? It is described above in 'Requirements' section.
If you don't pass `SERVER_CERT` or `SERVER_PRIVATEKEY` server will run in HTTP mode (unencrypted!) which is not recommended for production!

If you want to run CLI commands in docker, from the host machine type:
```bash
docker exec -it ddns-aws /ddns-aws <command> 
```


## Usage
CLI can be used to configure service, enable or disable managed domains, update IPs, etc.

When updating IPs and managing records, service does not need to be restarted, config is reloaded automatically. Hoeverer, if you want to change Server settings or AWS credentials, it is recommended to restart the server.

### all command line options
```bash
Usage: ddns-aws <command>

Flags:
  -h, --help    Show context-sensitive help.

Manage records:
  records list <zone-id>
    List all A records in the hosted zone

  records activate <zone-id> <resource-name>
    Activate record for DDNS updates

  records deactivate <resource-name>
    Deactivate record for DDNS updates

  records update <resource-name> [<ip>]
    Manually update IP address of the record

Manage hosted zones:
  zones list

Server: Run DDNS server
  server run

Run "ddns-aws <command> --help" for more information on a command.
```

If you want to run CLI commands in docker, from the host machine type:
```bash
docker exec -it ddns-aws /ddns-aws <command> 
```

## examples
Note, if you only want to run ddns-aws as a regular user, you can move config file from `/etc` to the `$HOME` directory of that user. Sudo will not be needed then.

Usually, first command will be:
```bash
ddns-aws zones list
```
This will give you list of all your hosted zones and their IDs. Copy one of the zone IDs and list all A records:
```bash
ddns-aws records list <zoneID>
```
Note that all domains in the list are disabled by default(empty [ ]) and DDNS-AWS will refuse to update their IP records until activated. Do this by:
```bash
sudo ddns-aws records activate <zoneID> <my_domain_with_trailing_dot>
```
Check the list of domains again:
```bash
ddns-aws records list <zoneID>
```
... and your first domain should be active now. Let's update IP for it manually:
```bash
sudo ddns-aws records update <my_domain_with_trailing_dot> 8.8.8.8
```
You can try skipping 8.8.8.8 above. DDNS-AWS will try to figure out your IP, but if it fails, specify it manually like above. Then check status again:
```bash
ddns-aws records list <zoneID>
```
... and new IP should be displayed.

To manually start the server:
```bash
ddns-aws server run
```
But before running it, check your config file and server configuration.
For production, it is strongly recommended to run DDNS-AWS as a service or as a Docker imagem, AND with server certificates configured in your config (which enables HTTPS).
## Contributing

This is a project used to learn Go (I'm also using it for my own domains), so any help is appreciated. But my goal is to keep it simple and small so for bigger changes I guess forking is a better option.


## License

[MIT](https://choosealicense.com/licenses/mit/)


