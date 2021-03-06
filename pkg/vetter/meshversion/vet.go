/*
Copyright 2017 Aspen Mesh Authors.

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

// Package meshversion vets the versions of Istio components, Sidecar proxy
// and generates notes on version mismatch.
package meshversion

import (
	"errors"

	apiv1 "github.com/aspenmesh/istio-vet/api/v1"
	"github.com/aspenmesh/istio-vet/pkg/vetter"
	"github.com/aspenmesh/istio-vet/pkg/vetter/util"
	"github.com/golang/glog"
	"k8s.io/client-go/listers/core/v1"
)

const (
	vetterID                       = "MeshVersion"
	latestTag                      = "latest"
	istioComponentMismatchNoteType = "istio-component-mismatch"
	istioComponentMismatchSummary  = "Mismatched istio component versions - ${component_name}"
	istioComponentMismatchMsg      = "Istio component ${component_name} is running version ${component_version}" +
		" but your environment is running istio version ${istio_version}." +
		" Consider upgrading the component ${component_name} "
	sidecarMismatchNoteType = "sidecar-image-mismatch"
	sidecarMismatchSummary  = "Mismatched sidecar image - ${pod_name}"
	sidecarMismatchMsg      = "The pod ${pod_name} in namespace ${namespace}" +
		" is running with sidecar proxy image ${sidecar_image}" +
		" but your environment is injecting ${inject_sidecar_image} for" +
		" new workloads. Consider upgrading the sidecar proxy in the pod."
	initMismatchNoteType = "init-image-mismatch"
	initMismatchSummary  = "Mismatched istio-init image - ${pod_name}"
	initMismatchMsg      = "The pod ${pod_name} in namespace ${namespace}" +
		" is running with istio-init image ${init_image}" +
		" but your environment is injecting ${inject_init_image} for" +
		" new workloads. Consider upgrading the istio-init container in the pod."
	missingVersionNoteType    = "missing-version"
	missingVersionNoteSummary = "Missing version information"
	missingVersionNoteMsg     = "Cannot determine mesh version"
)

// MeshVersion implements Vetter interface
type MeshVersion struct {
	podLister v1.PodLister
	cmLister  v1.ConfigMapLister
	nsLister  v1.NamespaceLister
}

type injectImages struct {
	Init    string
	Sidecar string
}

func getInjectImages(cmLister v1.ConfigMapLister) (injectImages, error) {
	sic, err := util.GetInitializerSidecarSpec(cmLister)
	if err != nil {
		return injectImages{}, err
	}
	if len(sic.InitContainers) > 0 && len(sic.Containers) > 0 {
		return injectImages{
			Init:    sic.InitContainers[0].Image,
			Sidecar: sic.Containers[0].Image,
		}, nil
	} else {
		errStr := "Failed to get inject images"
		glog.Error(errStr)
		return injectImages{}, errors.New(errStr)
	}
}

// The istio-inject ConfigMap has the sidecar & init images that should be
// injected into all new deployments, daemonsets, ....  If that doesn't match
// the images that are injected, emit a warning.
func (m *MeshVersion) vetInjectedImages() ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}
	injectImages, err := getInjectImages(m.cmLister)
	if err != nil {
		if n := util.IstioInitializerDisabledNote(err.Error(), vetterID,
			sidecarMismatchNoteType); n != nil {
			notes = append(notes, n)
		}
		return notes, nil
	}

	pods, err := util.ListPodsInMesh(m.nsLister, m.cmLister, m.podLister)
	if err != nil {
		// If err != nil when getting pod data, the lower-level error has already
		// been logged and handled.
		return nil, err
	}

	for _, p := range pods {
		sidecarImage, err := util.Image(util.IstioProxyContainerName, p.Spec)
		if err == nil && sidecarImage != injectImages.Sidecar {
			notes = append(notes, &apiv1.Note{
				Type:    sidecarMismatchNoteType,
				Summary: sidecarMismatchSummary,
				Msg:     sidecarMismatchMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"pod_name":             p.Name,
					"namespace":            p.Namespace,
					"sidecar_image":        sidecarImage,
					"inject_sidecar_image": injectImages.Sidecar}})
		}

		initImage, err := util.InitImage(util.IstioInitContainerName, p.Spec)
		if err == nil && initImage != injectImages.Init {
			notes = append(notes, &apiv1.Note{
				Type:    initMismatchNoteType,
				Summary: initMismatchSummary,
				Msg:     initMismatchMsg,
				Level:   apiv1.NoteLevel_WARNING,
				Attr: map[string]string{
					"pod_name":          p.Name,
					"namespace":         p.Namespace,
					"init_image":        initImage,
					"inject_init_image": injectImages.Init}})
		}
	}
	return notes, nil
}

// Vet returns the list of generated notes
func (m *MeshVersion) Vet() ([]*apiv1.Note, error) {
	notes := []*apiv1.Note{}

	injectedNotes, err := m.vetInjectedImages()
	if err == nil {
		notes = append(notes, injectedNotes...)
	}

	for i := range notes {
		notes[i].Id = util.ComputeID(notes[i])
	}

	return notes, nil
}

// Info returns information about the vetter
func (m *MeshVersion) Info() *apiv1.Info {
	return &apiv1.Info{Id: vetterID, Version: "0.1.0"}
}

// NewVetter returns "MeshVersion" which implements Vetter Interface
func NewVetter(factory vetter.ResourceListGetter) *MeshVersion {
	return &MeshVersion{
		podLister: factory.K8s().Core().V1().Pods().Lister(),
		cmLister:  factory.K8s().Core().V1().ConfigMaps().Lister(),
		nsLister:  factory.K8s().Core().V1().Namespaces().Lister(),
	}
}
