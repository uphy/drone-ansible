# drone-ansible

Use the Drone plugin to provision with ansible.
The following parameters are used to configure this plugin:

* `inventory` - define the inventory file (default: staging)
* `inventories` - define multiple inventory files to deploy
* `inventory_path`-  define the path in the project for ansible inventory files (default: provisioning/inventory)
* `playbook` - define the playbook file (default: provisioning/provision.yml)
* `ssh_key` - define the ssh_key to use for connecting to hosts
* `ssh_user` - define the ssh_user to specify the SSH login user name
* `ssh_passphrase` - define the passphrase for the SSH private key

The following is a sample configuration in your .drone.yml file:

```yaml
pipeline:
  deploy-staging:
    image: uphy/drone-ansible:2
    inventory: staging
    secrets: [ ssh_key ]
    when:
      branch: master
```

```yaml
pipeline:
  deploy-staging:
    image: uphy/drone-ansible:2
    inventories: [ staging, staging_2 ]
    secrets: [ ssh_key ]
    when:
      branch: master
```

To add the ssh key use drone secrets via the cli

```
drone secret add \
  -repository user/repo \
  -image uphy/drone-ansible \
  -name ssh_key \
  -value @Path/to/.ssh/id_rsa
```

Exposed Drone variables to Ansible which can be used in any playbook:

```
commit_tag -> DRONE_TAG
commit_sha -> DRONE_COMMIT_SHA
```

## Build

Build the binary with the following commands:

```
go build
go test
```

## Docker

Build the docker image with the following commands:

```
docker build --rm=true -t uphy/drone-ansible:2 .
```

Please note incorrectly building the image for the correct x64 linux and with
GCO disabled will result in an error when running the Docker image:

```
docker: Error response from daemon: Container command
'/bin/drone-ansible' not found or does not exist..
```

## Local usage

Execute from a project directory:

```
docker run --rm=true \
  -e PLUGIN_SSH_KEY=${SSH_KEY} \
  -e DRONE_WORKSPACE=/go/src/github.com/username/test \
  -v $(pwd):/go/src/github.com/username/test \
  -w /go/src/github.com/username/test \
  uphy/drone-ansible:2
```
