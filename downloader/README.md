# Downloader

This program is based on a skeleton kubebuilder. See https://book.kubebuilder.io/.
The kubebuilder version used is:

* `Version: version.Version{KubeBuilderVersion:"2.3.1", KubernetesVendor:"1.16.4", GitCommit:"8b53abeb4280186e494b726edf8f54ca7aa64a49", BuildDate:"2020-03-26T16:42:00Z", GoOs:"unknown", GoArch:"unknown"}`

The skeleton was created with the following commands:
* `kubebuilder init --domain kontain.app`
* `kubebuilder create api --group build --kind Image --version v1` (generate crd and controller)

The files that have actual changes in them are:

* `api/v1/image_types.go`
* `controllers/image_controller.go`

This controller runs on every pod with the faas-server.
