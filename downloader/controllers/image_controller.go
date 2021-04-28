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

	"github.com/google/uuid"

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
	functionDownloadDir := "/kontain/downloads/" + uuid.New().String()
	_, err := os.Stat(functionDownloadDir)
	if err == nil {
		image.Status.Message = "use existing"
		r.Log.Info("Use existing bundle", "name", req)
		return ctrl.Result{}, nil
	}

	// make a directory for the container image
	err = os.MkdirAll(functionDownloadDir, 0755)
	if err != nil {
		image.Status.Message = "mkdir failure"
		r.Log.Info("Make bundle director failed", "name", req, "err", err)
		return ctrl.Result{}, err
	}
	defer os.RemoveAll(functionDownloadDir)

	// Use skopeo to download image from registry
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	faasName := image.Spec.Image
	execCmd := exec.Command("/usr/bin/skopeo", "copy",
		"--src-daemon-host", dockerHost, "--src-cert-dir", dockerCertPath, "--insecure-policy",
		faasName, "oci:"+functionDownloadDir)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		image.Status.Message = "skopeo failure"
		r.Log.Info("execCmd failed (skopeo)", "err", err, "output", output)
		return ctrl.Result{}, err
	}
	image.Status.Message = "skopeo success"

	// Get the digest from index.json
	var digest string
	digest, err = r.OCIDigest(functionDownloadDir)
	if err != nil {
		r.Log.Info("OCIDigest failed", "err", err)
		return ctrl.Result{}, err
	}

	functionImageDir := "/kontain/images/" + req.NamespacedName.Namespace + "/" + req.NamespacedName.Name
	// Make a directory for the runtime bundle.
	err = os.MkdirAll(functionImageDir, 0755)
	if err != nil {
		r.Log.Info("MkdirAll failed", "err", err)
		return ctrl.Result{}, err
	}

	// build the runtime bundle using oci-image-tool
	execCmd = exec.Command("/usr/bin/oci-image-tool", "unpack",
		"--ref", "digest="+digest, functionDownloadDir, functionImageDir)
	output, err = execCmd.CombinedOutput()
	if err != nil {
		r.Log.Info("execCmd failed (oct-image-tool)", "err", err, "output", output)
		return ctrl.Result{}, err
	}

	// Remove download directoru
	err = os.RemoveAll(functionDownloadDir)
	if err != nil {
		r.Log.Info("RemoveAll failed", "err", err)
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ImageReconciler) DeleteFunction(req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("Delete record", "req", req)
	functionImageDir := "/kontain/images/" + req.NamespacedName.Namespace + "/" + req.NamespacedName.Name
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

func (r *ImageReconciler) OCIDigest(downloadDir string) (string, error) {
	type OCIManifest struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	}
	type OCIIndex struct {
		SchemaVersion int           `json:"schemaVersion"`
		Manifests     []OCIManifest `json:"manifests"`
	}

	data, err := ioutil.ReadFile(downloadDir + "/index.json")
	if err != nil {
		r.Log.Info("Readfile failed", "file", downloadDir+"/index.json", "err", err)
		return "", err
	}

	var ociIndex OCIIndex
	json.Unmarshal(data, &ociIndex)

	return ociIndex.Manifests[0].Digest, nil
}
