# quay-crd

[![](https://img.shields.io/badge/go-1.24.6-version?logo=go&color=rgb(0%2C%20126%2C%20198))]()
[![](https://img.shields.io/badge/Kubernetes-326CE5?logo=Kubernetes&logoColor=white)]()

Kubernetes Custom Resource Definition for manage [Quay registry](https://github.com/quay/quay)

## Features ✨

- **Organization CRDs** : Manage Quay's Organization by deploing/updating/deleting a Kubernetes CRD
- **Team CRDs** : Manage Quay's Organiztion Teams by deploing/updating/deleting a Kubernetes CRD
- **Group CRDs** : Set a group of members that can be reused in some Teams
- **Robot Account CRDs** : Manage Quay's Robot Account by deploing/updating/deleting a Kubernetes CRD
- **Repo CRDs**: Manage Quay's Repository (tag expiration, mirroring settings) by deploing/updating/deleting a Kubernetes CRD

## Get started :rocket:

### Requirements

* GO `>= 1.24.6`
* Docker `>= 17.03`
* [Kubebuilder](https://book.kubebuilder.io/quick-start.html)
* Make 
* Kubernetes cluster `>= 1.11.3`
  * Kubectl `>= 1.11.3`
* Quay registry
  * API token from a superuser (of reference)

### Download dependencies

```sh
go mod download
```

### Create a new CRD (API)

```sh
kubebuilder create api --version v1alpha --kind <kind of the new api>
```

### Build and push controller image

```sh
BUILDX_NO_DEFAULT_ATTESTATIONS=1 make docker-build docker-push IMG=<some-registry>/quay-crd-controller:tag
```

### Install the CRDs into the cluster

```sh
make install
```

> [!NOTE]
> This command, deploy the generated manifest : CRDs, RBAC, Service Account...
> The default namespace for these resources is `quay-crd-system`


### Deploy the controller

```sh
make deploy IMG=<some-registry>/quay-crd-controller:tag
```

> [!NOTE]
> This command, deploy the Deployment of the controller into the default namespace `quay-crd-system`

### Deploy sample resources

```sh
kubectl apply -k config/samples/
```

See also [build-instructions.md](./build-instructions.md)

## Documentation 📚

You can find the CRDs documentation in the [wiki](https://github.com/BenjaminFourmaux/quay-crd/wiki)

## Version

[![](https://badgen.net/github/tag/BenjaminFourmaux/quay-crd?cache=600)](https://github.com/BenjaminFourmaux/quay-crd/tags) [![](https://badgen.net/github/release/BenjaminFourmaux/quay-crd}?cache=600)](https://github.com/BenjaminFourmaux/quay-crd/releases)

- **v1alpha** :  First version of CRDs with Organization, Team, Group, Robot Account and Repo

## Contributors 👪

[![](https://badgen.net/github/contributors/BenjaminFourmaux/quay-crd)](https://github.com/BenjaminFourmaux/quay-crd/graphs/contributors)

- :crown: [Benjamin Fourmaux](https://github.com/BenjaminFourmaux)

## Licence ⚖️

All files on this project is under [**Apache License v2**](https://www.apache.org/licenses/LICENSE-2.0).
You can:
- Reuse the code 
- Modified the code
- Build the code

You must **Mention** the © Copyright if you use and modified code for your own profit. Thank you

---

© 2026 - Benjamin Fourmaux
