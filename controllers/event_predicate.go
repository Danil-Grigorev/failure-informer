package controllers

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type EventPredicate struct {
	predicate.Funcs
}

func (r EventPredicate) Create(e event.CreateEvent) bool {
	event, cast := e.Object.(*corev1.Event)
	if cast {
		return event.InvolvedObject.Kind == "Pod"
	}
	return false
}

func (r EventPredicate) Update(e event.UpdateEvent) bool {
	return false
}

func (r EventPredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (r EventPredicate) Generic(e event.GenericEvent) bool {
	return false
}
