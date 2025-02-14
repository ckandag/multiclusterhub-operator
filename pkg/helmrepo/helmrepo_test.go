// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package helmrepo

import (
	"reflect"
	"testing"

	operatorsv1 "github.com/stolostron/multiclusterhub-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeployment(t *testing.T) {
	empty := &operatorsv1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec: operatorsv1.MultiClusterHubSpec{
			ImagePullSecret: "",
		},
	}
	ovr := map[string]string{}

	t.Run("MCH with empty fields", func(t *testing.T) {
		_ = Deployment(empty, ovr)
	})

	essentialsOnly := &operatorsv1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec:       operatorsv1.MultiClusterHubSpec{},
	}
	t.Run("MCH with only required values", func(t *testing.T) {
		_ = Deployment(essentialsOnly, ovr)
	})
}

func TestService(t *testing.T) {
	mch := &operatorsv1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testName",
			Namespace: "testNS",
		},
	}

	t.Run("Create service", func(t *testing.T) {
		s := Service(mch)
		if ns := s.Namespace; ns != "testNS" {
			t.Errorf("expected namespace %s, got %s", "testNS", ns)
		}
		if ref := s.GetOwnerReferences(); ref[0].Name != "testName" {
			t.Errorf("expected ownerReference %s, got %s", "testName", ref[0].Name)
		}
	})
}

func TestValidateDeployment(t *testing.T) {
	mch := &operatorsv1.MultiClusterHub{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
		Spec: operatorsv1.MultiClusterHubSpec{
			ImagePullSecret: "test",
			NodeSelector: map[string]string{
				"test": "test",
			},
		},
	}
	ovr := map[string]string{}

	// 1. Valid mch
	dep := Deployment(mch, ovr)

	// 2. Modified ImagePullSecret
	dep1 := dep.DeepCopy()
	dep1.Spec.Template.Spec.ImagePullSecrets = nil

	// 3. Modified image
	dep2 := dep.DeepCopy()
	dep2.Spec.Template.Spec.Containers[0].Image = "differentImage"

	// 4. Modified pullPolicy
	dep3 := dep.DeepCopy()
	dep3.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullNever

	// 5. Modified NodeSelector
	dep4 := dep.DeepCopy()
	dep4.Spec.Template.Spec.NodeSelector = nil

	// 6. Modified Tolerations
	dep5 := dep.DeepCopy()
	dep5.Spec.Template.Spec.Tolerations = nil

	// 7. Missing deployment labels
	dep6 := dep.DeepCopy()
	dep6.Labels = selectorLabels()

	// 8. Missing deployment labels
	dep7 := dep.DeepCopy()
	dep7.Spec.Template.Labels = selectorLabels()

	type args struct {
		m   *operatorsv1.MultiClusterHub
		dep *appsv1.Deployment
	}
	tests := []struct {
		name  string
		args  args
		want  *appsv1.Deployment
		want1 bool
	}{
		{
			name:  "Valid Deployment",
			args:  args{mch, dep},
			want:  dep,
			want1: false,
		},
		{
			name:  "Modified ImagePullSecret",
			args:  args{mch, dep1},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified Image",
			args:  args{mch, dep2},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified PullPolicy",
			args:  args{mch, dep3},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified NodeSelector",
			args:  args{mch, dep4},
			want:  dep,
			want1: true,
		},
		{
			name:  "Modified Tolerations",
			args:  args{mch, dep5},
			want:  dep,
			want1: true,
		},
		{
			name:  "Missing deployment labels",
			args:  args{mch, dep6},
			want:  dep,
			want1: true,
		},
		{
			name:  "Missing pod labels",
			args:  args{mch, dep7},
			want:  dep,
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ValidateDeployment(tt.args.m, ovr, dep, tt.args.dep)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateDeployment() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ValidateDeployment() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
