package controllers

import (
	"context"
	"time"

	buildv1 "faas.kontain.app/api/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("the image downloader", func() {
	Context("a new image record appears", func() {
		When("the image does not exist", func() {
			It("fails", func() {
				const (
					FunctionName      = "test-function"
					FunctionNamespace = "test-function-namespace"
					ImageName         = "docker://docker.io/muthatkontain/hello:a52ced9a4c343acf4a57f8fa9c40c5d57676b620a7d041bcd8b53e47b311ff80"

					timeout  = time.Second * 30
					duration = time.Second * 10
					interval = time.Millisecond * 250
				)

				ctx := context.Background()
				image := &buildv1.Image{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "build.kontain.app/v1",
						Kind:       "Image",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      FunctionName,
						Namespace: FunctionNamespace,
					},
					Spec: buildv1.ImageSpec{
						Image: ImageName,
					},
				}
				Expect(k8sClient.Create(ctx, image)).Should(Succeed())

				imageLookupKey := types.NamespacedName{Name: FunctionName, Namespace: FunctionNamespace}
				createdImage := &buildv1.Image{}

				// We'll need to retry getting this newly created Image, given that creation may not immediately happen.
				Eventually(func() bool {
					err := k8sClient.Get(ctx, imageLookupKey, createdImage)
					if err != nil {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())
				// Let's make sure our Schedule string value was properly converted/handled.
				Expect(createdImage.Spec.Image).Should(Equal(ImageName))

				Expect(k8sClient.Delete(ctx, image)).Should(Succeed())
			})
		})
	})
})
