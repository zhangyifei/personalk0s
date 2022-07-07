Ericsson Web Services CLI

# Development Environment
https://gitlab.rnd.gic.ericsson.se/raket/eke
Make the repo readable to internal Ericsson users and do it community way.
- Source code and procedures in README.md
- Setup Gitlab CI
- Release tag and release notes
- Binary publish to /infra/ews/www/release/vx.y.z-eke.n and ARM https://arm.seli.gic.ericsson.se/artifactory/proj-ews-generic/eke

# Prerequisite
User must has an EWS account with ssh public key uploaded and user agreement signed via https://ews.rnd.gic.ericsson.se/p.php.

If an idm group is to be used in Kubernetes authentication, it must first be imported to ews via https://ews.rnd.gic.ericsson.se/g.php by any member in that group.

Both of above are one time action priort to using EWS service.

# Installation

Download eke binary and mv it to user PATH
```
# Linux
curl -O https://www.rnd.gic.ericsson.se/vx.y.z-eke.n/bin/linux/amd64/eke
# MacOS
curl -O https://www.rnd.gic.ericsson.se/vx.y.z-eke.n/bin/darwin/amd64/eke
# Windows
curl -O https://www.rnd.gic.ericsson.se/vx.y.z-eke.n/bin/windows/amd64/eke.exe

# For Linux or MacOS
chmod 755 eke
# Either move it to shared user local bin dir
sudo mv eke /usr/local/bin
# Or move it user's home bin dir
mkdir -p ~/bin
mv eke ~/bin
export PATH=$PATH:~/bin

# For Windows, the user's client may vary, e.g. CMD, Powershell, WSL.
# Just do the necessary same to make sure eke.exe can be found in user PATH.
```

If not done before, download kubectl binary and make it in user PATH
```
which kubectl
/usr/local/bin/kubectl
```
# Kubernetes Client Key and Certificate
A Kubernetes client key and certificate pair will be generated for all EWS users with 7 days valid period and will be automatically renewed upon expiration.

User can get or force renew the client key and certificate via EWS CLI below or GUI at https://ews.rnd.gic.ericsson.se/p.php. User's Ericsson signum and password will be asked for validation.

```
# Get EWS client key and certificate
eke ckc get
# if user has client key/certificate cached in local already, it will just display that stored in local.
# otherwise, it will fetch it from remote and user's signum and password will be asked.

# Renew EWS client key and certificate
eke ckc renew
signum: ?
password: ?
# renewed client key and certificate will be cached in local

# user can also specify signum and password using following options to avoid interactive mode
--userid/-u <ericsson signum>
--password/-p <user password>
```

# Kubeconfig
User can either use default kubeconfig file at ~/.kube/config or specify it via --kubeconfig.

```
# init the kubeconfig file and ews config dir ~/.eke/
eke kubeconfig init <cluster name> [--static] [--kubeconfig <filename>]
signum: ?
password: ?

# user can also specify signum and password using following options to avoid interactive mode
--userid/-u <ericsson signum>
--password/-p <user password>

# clean up kubeconfig file and ews config dir ~/.eke/
eke kubeconfig reset [--kubeconfig <filename>]

# get kubeconfig file content
eke kubeconfig get [--kubeconfig <filename>]
```

By default, the generated kubeconfig file will use [client go credential plugins](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins) to call eke client for user authentication. It will check and use the local cached client key and certificate pair for Kubernetes authentication. If the certificate is expired, it will ask for user signum and password to authenticate with Ericsson LDAP and automatically generate/cache a new pair for use. The user context in kubeconfig file will be something like below:
```
apiVersion: v1
kind: Config
users:
- name: eacehij
  user:
    exec:
      command: "eke"
      apiVersion: "client.authentication.k8s.io/v1beta1"
      args:
      - "kubeconfig"
      - "auth"
```

For any use case that the client go credential plugins does not work, e.g. bypass kubectl to talk to apiserver in other means, user can generate a static kubeconfig file, which will embed the client key and cer without invoking eke client. However, once the client certificate expires, it will not be automatically renewed and user has to generate the kubeconfig file again. 
```
eke kubeconfig init <cluster name> --static

# the user context will be something like below:
apiVersion: v1
kind: Config
users:
- name: eacehij
  user:
    client-certificate-data: <base64 encoding of client certificate>
    client-key-data: <base64 encoding of client key>
``` 


# Use kubectl or helm client
```
kubectl [--kubeconfig <filename>] command options
helm [--kubeconfig <filename>] command options
```
# Build
```
# build for current platform
./build/build.sh

# build for cross platform
CROSS=1 ./build/build.sh

ls ./build/bin
```
also , can use make to do build, you can use TARGET_OS to define OS. default is linux
```
#build for linux
make

#build for windows
make TARGET_OS=windows

#build for darwin
make TARGET_OS=darwin

#build for 3 platform together
make all
```

