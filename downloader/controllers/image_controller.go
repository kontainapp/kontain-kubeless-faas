/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	buildv1 "faas.kontain.app/api/v1"
)

// ImageReconciler reconciles a Image object
type ImageReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=build.kontain.app,resources=images,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=build.kontain.app,resources=images/status,verbs=get;update;patch

func (r *ImageReconciler) OCIImagePath(namespace string, function string) string {
	return "/kontain/oci-images/" + namespace + "/" + function
}

func (r *ImageReconciler) OCIRuntimeBundlePath(namespace string, function string) string {
	return "/kontain/oci-runtime-bundles/" + namespace + "/" + function
}

func (r *ImageReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("image", req.NamespacedName)

	var image buildv1.Image
	if err := r.Get(ctx, req.NamespacedName, &image); err != nil {
		// Interpret NotFound as a delete.
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch Image record")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return r.DeleteFunction(req)
	}
	return r.CreateFunction(req, image)
}

func (r *ImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&buildv1.Image{}).
		Complete(r)
}

func (r *ImageReconciler) CreateFunction(req ctrl.Request, image buildv1.Image) (ctrl.Result, error) {
	r.Log.Info("New function image", "req", req)

	// If we have a copy of the function container, use it
	ociImageDir := r.OCIImagePath(req.NamespacedName.Namespace, req.NamespacedName.Name)
	ociRuntimeBundleDir := r.OCIRuntimeBundlePath(req.NamespacedName.Namespace, req.NamespacedName.Name)

	_, err := os.Stat(ociRuntimeBundleDir)
	if err == nil {
		image.Status.Message = "use existing"
		r.Log.Info("Use existing bundle", "name", ociRuntimeBundleDir)
		return ctrl.Result{}, nil
	}

	// If OCI image is already there, use it
	_, err = os.Stat(ociImageDir)
	if err != nil {
		/*
			if err != fs.ErrNotExist {
				return ctrl.Result{}, err
			}
		*/

		// Get OCI Image
		err = os.MkdirAll(ociImageDir, 0755)
		if err != nil {
			image.Status.Message = "mkdir failure"
			r.Log.Info("Make bundle director failed", "name", req, "err", err)
			return ctrl.Result{}, err
		}
		// defer os.RemoveAll(ociImageDir)

		// Use skopeo to download image from registry
		faasName := image.Spec.Image
		/*
			dockerHost := os.Getenv("DOCKER_HOST")
			dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
			execCmd := exec.Command("/usr/bin/skopeo", "copy",
				"--src-daemon-host", dockerHost, "--src-cert-dir", dockerCertPath, "--insecure-policy",
				faasName, "oci:"+ociImageDir)
		*/
		execCmd := exec.Command("/usr/bin/skopeo", "copy", "--insecure-policy",
			faasName, "oci:"+ociImageDir)
		output, err := execCmd.CombinedOutput()
		if err != nil {
			image.Status.Message = "skopeo failure"
			r.Log.Info("execCmd failed (skopeo)", "err", err, "output", output)
			return ctrl.Result{}, err
		}
		image.Status.Message = "skopeo success"
	}

	// Get the digest from index.json
	var digest string
	digest, err = r.OCIDigest(ociImageDir)
	if err != nil {
		r.Log.Info("OCIDigest failed", "err", err)
		return ctrl.Result{}, err
	}
	r.Log.Info("image manifest:", "digest", digest)

	// TODO: unpack into someplace else and rename(2) when done. Provents server from seeing half built things
	functionImageDir := "/kontain/oci-runtime-bundles/" + req.NamespacedName.Namespace + "/" + req.NamespacedName.Name
	// Make a directory for the runtime bundle.
	err = os.MkdirAll(functionImageDir+"/rootfs", 0755)
	if err != nil {
		r.Log.Info("MkdirAll failed", "err", err)
		return ctrl.Result{}, err
	}

	{

		// build the runtime bundle using oci-image-tool
		execCmd := exec.Command("/usr/bin/oci-image-tool", "unpack",
			"--ref", "digest="+digest, ociImageDir, functionImageDir+"/rootfs")
		output, err := execCmd.CombinedOutput()
		if err != nil {
			r.Log.Info("execCmd failed (oct-image-tool)", "err", err, "output", output)
			return ctrl.Result{}, err
		}
	}

	err = r.OCIConfigCopy(ociImageDir, digest, functionImageDir)
	if err != nil {
		r.Log.Info("OCIConfigCopy failed", "err", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ImageReconciler) DeleteFunction(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Delete record", "req", req)
	functionImageDir := "/kontain/oci-runtime-bundles/" + req.NamespacedName.Namespace + "/" + req.NamespacedName.Name
	_, err := os.Stat(functionImageDir)
	if err != nil {
		r.Log.Info("Bundle does not exist", "name", req)
		return ctrl.Result{}, err
	}
	if err = os.RemoveAll(functionImageDir); err != nil {
		r.Log.Info("Bundle Remove failed", "name", req, "err", err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

/*
 * These types are used to decode the metadata for an OCI Image Layout, starting with index.json
 * OCI Image decode. See https://github.com/opencontainers/image-spec/blob/master/spec.md
 */

// OCI index.json
type OCIIndex struct {
	SchemaVersion int                `json:"schemaVersion"`
	Manifests     []OCIIndexManifest `json:"manifests"`
	Annotations   map[string]string  `json:"annotations,omitempty"`
}
type OCIIndexManifest struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

// OCI image manifest - blob, referenced by index.json
type OCIImageManifest struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Config        OCIImageDescriptor   `json:"config"`
	Layers        []OCIImageDescriptor `json:"layer"`
	Annotations   map[string]string    `json:"annotations,omitempty"`
}

// OCI Content Descriptor for items in OCIImageManifest
type OCIImageDescriptor struct {
	MediaType   string            `json:"mediaType"`
	Digest      string            `json:"digest"`
	Size        int               `json:"size"`
	Urls        int               `json:"urls,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// OCI Image Configurationa - blob, pointed to by OCIImageManifest.Config
type OCIImageConfiguration struct {
	Created      string             `json:"created,omitempty"`
	Author       string             `json:"author,omitempty"`
	Architecture string             `json:"architecture"`
	Os           string             `json:"os"`
	Config       OCIImageExecConfig `json:"config"`
}

//
type OCIImageExecConfig struct {
	User         string              `json:"User,omitempty"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"`
	Env          []string            `json:"Env,omitempty"`
	Entrypoint   []string            `json:"ENtrypoint,omitempty"`
	Cmd          []string            `Entrypoint:"Cmd,omitempty"`
	Volumes      map[string]struct{} `json:"Volumesexposedports,omitempty"`
	WorkingDir   string              `json:"WorkingDir,omitempty"`
	Labels       map[string]string   `jsoon:"Labels,omitempty"`
	StopSignal   string              `json:"StopSignal,omitempty"`
}

func (r *ImageReconciler) OCIDigest(downloadDir string) (string, error) {

	data, err := ioutil.ReadFile(downloadDir + "/index.json")
	if err != nil {
		r.Log.Info("Readfile failed", "file", downloadDir+"/index.json", "err", err)
		return "", err
	}

	var ociIndex OCIIndex
	json.Unmarshal(data, &ociIndex)

	return ociIndex.Manifests[0].Digest, nil
}

func (r *ImageReconciler) OCIConfigCopy(downloadDir string, digest string, imageDir string) error {
	digest_arr := strings.Split(digest, ":")
	data, err := ioutil.ReadFile(downloadDir + "/blobs/" + digest_arr[0] + "/" + digest_arr[1])
	if err != nil {
		r.Log.Info("Readfile failed", "file", downloadDir+"/index.json", "err", err)
		return err
	}

	var ociManifest OCIImageManifest
	json.Unmarshal(data, &ociManifest)
	r.Log.Info("config", "desc", ociManifest.Config)

	digest_arr = strings.Split(ociManifest.Config.Digest, ":")
	data, err = ioutil.ReadFile(downloadDir + "/blobs/" + digest_arr[0] + "/" + digest_arr[1])
	if err != nil {
		r.Log.Info("Readfile failed", "file", downloadDir+"/index.json", "err", err)
		return err
	}
	err = ioutil.WriteFile(imageDir+"/oci-config.json", data, 0444)
	return err
}
