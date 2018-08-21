package meshversion

import (
	"fmt"
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeSIS(img1, img2 string) util.SidecarInjectionSpec {
	return util.SidecarInjectionSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Image: img1,
			},
		},
		Containers: []corev1.Container{
			corev1.Container{
				Image: img2,
			},
		},
	}
}
func halfSIS(img1 string) util.SidecarInjectionSpec {
	return util.SidecarInjectionSpec{
		InitContainers: []corev1.Container{
			corev1.Container{
				Image: img1,
			},
		},
	}
}
func pod(name, namespace, scImage, initImage string) *corev1.Pod {

	//add additional containers with bogus data for realistic-ness
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Name:  "istio-proxy",
					Image: scImage,
				},
			},
			InitContainers: []corev1.Container{
				corev1.Container{
					Name:  "istio-init",
					Image: initImage,
				},
			},
		},
	}
}

func sortNotes(notes []*apiv1.Note) []*apiv1.Note {
	// don't need the if/else to deal with non-existant attrs
	sort.Slice(notes, func(i, j int) bool {
		if notes[i].Attr["namespace"] != notes[j].Attr["namespace"] {
			return notes[i].Attr["namespace"] < notes[j].Attr["namespace"]
		}
		if notes[i].Attr["pod_name"] != notes[j].Attr["pod_name"] {
			return notes[i].Attr["pod_name"] < notes[j].Attr["pod_name"]
		}
		if len(notes[i].Attr["init_image"]) == 0 {
			if notes[i].Attr["init_image"] != notes[j].Attr["init_image"] {
				return notes[i].Attr["init_image"] < notes[j].Attr["init_image"]
			}
			if notes[i].Attr["inject_init_image"] != notes[j].Attr["inject_init_image"] {
				return notes[i].Attr["inject_init_image"] < notes[j].Attr["inject_init_image"]
			}

		} else {
			if notes[i].Attr["sidecar_image"] != notes[j].Attr["sidecar_image"] {
				return notes[i].Attr["sidecar_image"] < notes[j].Attr["sidecar_image"]
			}
			if notes[i].Attr["inject_sidecar_image"] != notes[j].Attr["inject_sidecar_image"] {
				return notes[i].Attr["inject_sidecar_image"] < notes[j].Attr["inject_sidecar_image"]
			}
		}
		return false
	})
	return notes
}

var _ = Describe("Meshversion", func() {

	Describe("Meshversion can write Notes", func() {
		It("returns the right notes", func() {
			pods := []*corev1.Pod{}
			imagedot8 := "docker.io/istio/proxy_init:0.8.0"
			image1dot0 := "docker.io/istio/proxy_init:1.0.0"

			// pod(name, namespace, scImage, initImage)
			a := pod("name1", "namespace1", image1dot0, image1dot0) //make no note
			b := pod("name2", "namespace1", imagedot8, imagedot8)   //make 2 notes:
			c := pod("name3", "namespace1", image1dot0, imagedot8)  //make 1 note
			d := pod("name4", "namespace1", imagedot8, image1dot0)  //make 1 note

			pods = append(pods, a, b, c, d)

			iImages := injectImages{
				Init:    imagedot8,
				Sidecar: imagedot8,
			}

			notes := vetPods(pods, iImages)
			sortNotes(notes)
			fmt.Printf("\nnotes[0] %v | %v | %v | %v ", notes[0].Attr["namespace"], notes[0].Attr["pod_name"], notes[0].Attr["init_image"], notes[0].Attr["inject_init_image"])
			fmt.Printf("\nnotes[1] %v | %v | %v | %v ", notes[1].Attr["namespace"], notes[1].Attr["pod_name"], notes[1].Attr["sidecar_image"], notes[1].Attr["inject_sidecar_image"])
			fmt.Printf("\nnotes[2] %v | %v | %v | %v ", notes[2].Attr["namespace"], notes[2].Attr["pod_name"], notes[2].Attr["sidecar_image"], notes[2].Attr["inject_sidecar_image"])
			fmt.Printf("\nnotes[3] %v | %v | %v | %v ", notes[3].Attr["namespace"], notes[3].Attr["pod_name"], notes[3].Attr["init"], notes[3].Attr["inject_init"])

			Expect(len(notes)).To(Equal(4))

			Expect(notes[0].Attr["namespace"]).To(Equal("namespace1"))
			Expect(notes[0].Attr["pod_name"]).To(Equal("name1"))
			Expect(notes[0].Attr["init_image"]).To(Equal("docker.io/istio/proxy_init:1.0.0"))
			Expect(notes[0].Attr["inject_init_image"]).To(Equal("docker.io/istio/proxy_init:0.8.0"))
			Expect(notes[0].Attr["inject_sidecar_image"]).To(BeEmpty())
			Expect(notes[0].Msg).To(Equal(initMismatchMsg))

			Expect(notes[1].Attr["namespace"]).To(Equal("namespace1"))
			Expect(notes[1].Attr["pod_name"]).To(Equal("name1"))
			Expect(notes[1].Attr["sidecar_image"]).To(Equal("docker.io/istio/proxy_init:1.0.0"))
			Expect(notes[1].Attr["inject_sidecar_image"]).To(Equal("docker.io/istio/proxy_init:0.8.0"))
			Expect(notes[1].Attr["inject_init_image"]).To(BeEmpty())
			Expect(notes[1].Msg).To(Equal(sidecarMismatchMsg))

			Expect(notes[2].Attr["namespace"]).To(Equal("namespace1"))
			Expect(notes[2].Attr["pod_name"]).To(Equal("name3"))
			Expect(notes[2].Attr["sidecar_image"]).To(Equal("docker.io/istio/proxy_init:1.0.0"))
			Expect(notes[1].Attr["inject_sidecar_image"]).To(Equal("docker.io/istio/proxy_init:0.8.0"))
			Expect(notes[1].Attr["inject_init_image"]).To(BeEmpty())
			Expect(notes[1].Msg).To(Equal(sidecarMismatchMsg))

			Expect(notes[3].Attr["namespace"]).To(Equal("namespace1"))
			Expect(notes[3].Attr["pod_name"]).To(Equal("name4"))
			Expect(notes[3].Attr["init_image"]).To(Equal("docker.io/istio/proxy_init:1.0.0"))
			Expect(notes[0].Attr["inject_init_image"]).To(Equal("docker.io/istio/proxy_init:0.8.0"))
			Expect(notes[0].Attr["inject_sidecar_image"]).To(BeEmpty())
			Expect(notes[0].Msg).To(Equal(initMismatchMsg))

		})
	})

})
