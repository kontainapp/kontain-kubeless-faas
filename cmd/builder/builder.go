/*
Copyright (c) 2016-2017 Bitnami

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

// Kubeless controller binary.
//
// See github.com/kubeless/kubeless/tree/master/pkg/controller
package main

import (
	"flag"
	"kontain-faas/pkg/signals"

	"k8s.io/klog/v2"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// "k8s.io/sample-controller/pkg/signals"
	//	clientset "./pkg/generated/clientset/versioned"
	//	informers "./pkg/generated/informers/externalversions"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.Infof("hello world")

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	klog.Info(stopCh)

	/*
		cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
		if err != nil {
			klog.Fatalf("Error building kubeconfig: %s", err.Error())
		}

		kubeClient, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
		}

		exampleClient, err := clientset.NewForConfig(cfg)
		if err != nil {
			klog.Fatalf("Error building example clientset: %s", err.Error())
		}

		kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
		exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

		controller := NewController(kubeClient, exampleClient,
			kubeInformerFactory.Apps().V1().Deployments(),
			exampleInformerFactory.Samplecontroller().V1alpha1().Foos())

		// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
		// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
		kubeInformerFactory.Start(stopCh)
		exampleInformerFactory.Start(stopCh)

		if err = controller.Run(2, stopCh); err != nil {
			klog.Fatalf("Error running controller: %s", err.Error())
		}
	*/
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
