# Binderhub Setup

## Requirements

- Docker
- Kubernetes
- Kubectl CLI
- Helm

## Kubernetes Setup with Storage Class and Nginx Ingress for Baremetal Setup

### Docker Installation

Install docker

```shell
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates   curl gnupg lsb-release
sudo apt-get install -y docker-ce docker-ce-cli containerd.io
```

Configure docker for kubernetes 
```shell
cat <<EOF | sudo tee /etc/docker/daemon.json
{
    "exec-opts": ["native.cgroupdriver=systemd"],
    "log-driver": "json-file",
    "log-opts": {"max-size": "100m"},
    "storage-driver": "overlay2"
}
EOF
```

Restart docker and docker user group
```shell
sudo systemctl restart docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
```

### Kubeadm, Kubelet, Kubectl and Helm Installation

Install kubeadm, kubelet and kubectl

```shell
sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg   https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main"   | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

Ensure swap is disabled

```shell
swapon --show

sudo swapoff -a

sudo sed -i -e '/swap/d' /etc/fstab
```

Create cluster using kubeadm

```shell
sudo kubeadm init --pod-network-cidr=10.244.0.0/16
```

Configure kubectl

```shell
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

Untaint node to deploy cluster to our single node

```shell
kubectl taint nodes --all node-role.kubernetes.io/master-
```

Install flannel plugin for networking

```shell
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
```

Helm Install 

```shell
curl https://baltocdn.com/helm/signing.asc | sudo apt-key add -
sudo apt-get install apt-transport-https --yes
echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```

#### Storage Setup using OpenEBS

Install openebs using helm chart

```shell
helm repo add openebs https://openebs.github.io/charts
kubectl create namespace openebs
helm --namespace=openebs install openebs openebs/openebs
```

Configure default storage

```shell
kubectl get storageclass
kubectl patch storageclass openebs-hostpath -p '{"metadata": {"annotations": {"storageclass.kubernetes.io/is-default-class": "true"}}}'
```

Note: Storage class is required only in the case of baremetal setup

#### Nginx ingress with host network

Install nginx ingress

```shell
kubectl apply -f nginx-ingress-deploy.yaml
```

Run the following command with your server public IP

```shell
kubectl --namespace ingress-nginx patch svc ingress-nginx-controller -p '{"spec": {"type": "LoadBalancer", "externalIPs": ["<server-public-ip>"]}}'
```


### SSL Certificate Issuer

We are using certbot from letsencrypt for certificate issue for binder cluster

Install using below commands

```shell
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.6.1/cert-manager.yaml
kubectl apply -f binderhub-issuer.yaml
```

### Authentication Server

Any Oauth supported authorization server can be used to authenticate jupyter server. 

In this repo keycloak is used as example


### Container Registry

Binderhub build docker images using git repo, and it pushes to docker registry so that jupyterhub can launch user servers based on these images. 

In this example container registry is disabled but it is recommended to use in production environment.

Follow [this document](https://binderhub.readthedocs.io/en/latest/zero-to-binderhub/setup-registry.html) for configuring registry to binderhub


## Get Started

Refer the `secret.yaml` file in this repo. Replace apiToken and secretToken with new one. 

Command to generate tokens
```shell
openssl rand -hex 32
```

### Create namespace

Create `binder` namespace using kubectl

```shell
kubectl create ns binder
```

### Create all configmaps

We need 4 different config maps for running binder with persistent storage enabled

Creating config map for persistent storage
```shell
kubectl -n binder create -f persistent_config_configmap.yaml
```

Creating config map for scheduler config for token refresh
```shell
kubectl -n binder create -f scheduler_configmap.yaml
```

Creating supervisord config map for running the above scheduler in background
```shell
kubectl -n binder create -f supervisord_configmap.yaml
```

Creating jupyter config map for voila configuration (Currently this config is not being used)
```shell
kubectl -n binder create -f jupyter_config_configmap.yaml
```

### Config File

First step is to configure container registry follow the binderhub documentation.

Next config binderhub parameters. Enable user_registry if registry configured.
Provide image_prefix which can be stored in registry
Change auth_enabled accordingly.

`hub_url` can be updated later once jupyter_hub server is up. If you are configuring SSL then you can add it in this step itself.

Hub URL need to be replaced only in case of local setup. For prod setup it will be domain names.

```yaml
config:
  BinderHub:
    use_registry: false
    image_prefix: "local/prefix-"
    hub_url: http://10.103.164.198
    auth_enabled: true
```


Next step is to configure jupyterhub properties

`cull` configuration is disabled in the example. It is used to configure cleanup of jupyter server of inactive user. [Click here](https://zero-to-jupyterhub.readthedocs.io/en/latest/jupyterhub/customizing/user-management.html) for more info.

Note: Binderhub docs says to not cull authorized users.

```yaml
jupyterhub:
  ...
  hub:
    redirectToServer: false
    allowNamedServers: true
    namedServerLimitPerUser: 5
```
Named server config is to provide multiple servers for the same user. If its not provided then by default only one jupyter server is allowed per user.


Next step is to configure authentication flow for jupyterhub. Follow the example config for it. 

Below config is to add admin users. Its the usernames of the admins.
These users can use jupyterhub admin features.
```yaml
jupyterhub:
  ...
  hub:
    config:
      ...
      GenericOAuthenticator:
        ...
        admin_users:
          - adminuser1
      ....
```

Additionally in the example extraConfig is provided which is a python function. 

```yaml
jupyterhub:
 ....
 hub:
   ...
   extraConfig:
      authpasstoken: |
        from oauthenticator.generic import GenericOAuthenticator
        from tornado import gen
        class CustomKeycloakAuthenticator(GenericOAuthenticator):
          @gen.coroutine
          def pre_spawn_start(self, user, spawner):
            auth_state = yield user.get_auth_state()
            if not auth_state:
              return
            print(auth_state)
            spawner.environment['KC_ACCESS_TOKEN'] = auth_state['access_token']
            spawner.environment['KC_REFRESH_TOKEN'] = auth_state['refresh_token']
        c.JupyterHub.authenticator_class = CustomKeycloakAuthenticator
        c.Authenticator.enable_auth_state = True
```
It adds login access token and refresh token to environment variable. 

Note that this function runs only once i.e., during spawning container.

#### Database Config 

Database configuration is optional if not required then you can remove from the config file.
In production the database is shared with the sandbox backend hence it has to be configured.
By default jupyterhub has SQLite database in the shared storage space.

Commands to start the server

```shell
helm repo add jupyterhub https://jupyterhub.github.io/helm-chart
helm repo update
helm install binder jupyterhub/binderhub --version=0.2.0-n845.hcc57b24 --namespace=binder -f secret.yaml -f config-dev.yaml
```

For prod
```shell
helm install binder jupyterhub/binderhub --version=0.2.0-n845.hcc57b24 --namespace=binder -f secret.yaml -f config-prod.yaml
```

Replace the latest version from [here](https://jupyterhub.github.io/helm-chart/#development-releases-binderhub)


Both the commands below are for local dev setup

If you have not provided hub_url above run below command to get the address. (Skip this if its SSL based config)
```shell
kubectl --namespace=binder get svc proxy-public
```

Upgrade helm chart after updating config
```shell
helm upgrade binder jupyterhub/binderhub --version=0.2.0-n845.hcc57b24 --namespace=binder -f secret.yaml -f config-dev.yaml
```

You are good to go now. 

Open Binderhub URL which will redirect to jupyterhub URL to sign in. After that it will redirect back to binderhub homepage. Enter github repo it will build and launch the jupyterhub server.


### Resources

- For customization of binderhub UI template follow this documentation  https://binderhub.readthedocs.io/en/latest/customizing.html

- For SSL config follow this documenation - https://binderhub.readthedocs.io/en/latest/https.html

- Binderhub provides stream API which is used in its UI. If you wish to create your own UI for it you can use their API and build it - https://binderhub.readthedocs.io/en/latest/api.html

- Jupyterhub API docs
  1. https://jupyterhub.readthedocs.io/en/stable/reference/rest.html
  2. https://jupyterhub.readthedocs.io/en/stable/_static/rest-api/index.html

-  Limit resources in jupyterhub - https://zero-to-jupyterhub.readthedocs.io/en/latest/jupyterhub/customizing/user-resources.html
