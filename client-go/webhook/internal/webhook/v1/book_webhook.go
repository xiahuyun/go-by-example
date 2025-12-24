/*
Copyright 2025.

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

package v1

import (
	"context"
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	webappv1 "hxia.io/book/api/v1"
)

// nolint:unused
// log is for logging in this package.
var booklog = logf.Log.WithName("book-resource")

// SetupBookWebhookWithManager registers the webhook for Book in the manager.
func SetupBookWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&webappv1.Book{}).
		WithValidator(&BookCustomValidator{
			Client: mgr.GetClient(),
		}).
		WithDefaulter(&BookCustomDefaulter{}).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-webapp-hxia-io-v1-book,mutating=true,failurePolicy=fail,sideEffects=None,groups=webapp.hxia.io,resources=books,verbs=create;update,versions=v1,name=mbook-v1.kb.io,admissionReviewVersions=v1

// BookCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Book when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type BookCustomDefaulter struct {
	// TODO(user): Add more fields as needed for defaulting
}

var _ webhook.CustomDefaulter = &BookCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Book.
func (d *BookCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	book, ok := obj.(*webappv1.Book)

	if !ok {
		return fmt.Errorf("expected an Book object but got %T", obj)
	}
	booklog.Info("Defaulting for Book", "name", book.GetName())

	// TODO(user): fill in your defaulting logic.

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-webapp-hxia-io-v1-book,mutating=false,failurePolicy=fail,sideEffects=None,groups=webapp.hxia.io,resources=books,verbs=create;update;delete,versions=v1,name=vbook-v1.kb.io,admissionReviewVersions=v1

// BookCustomValidator struct is responsible for validating the Book resource
// when it is created, updated, or deleted.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type BookCustomValidator struct {
	client.Client
}

var _ webhook.CustomValidator = &BookCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type Book.
func (v *BookCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	book, ok := obj.(*webappv1.Book)
	if !ok {
		return nil, fmt.Errorf("expected a Book object but got %T", obj)
	}
	booklog.Info("Validation for Book upon creation", "name", book.GetName())

	name := book.Spec.Name
	if name == "" {
		return admission.Warnings{}, fmt.Errorf("book name is required")
	}

	if book.Spec.Name == "Python" {
		return admission.Warnings{}, fmt.Errorf("wrong book name %s", book.Spec.Name)
	}

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type Book.
func (v *BookCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	updatingBook, ok := newObj.(*webappv1.Book)
	if !ok {
		return nil, fmt.Errorf("expected a Book object for the newObj but got %T", newObj)
	}
	booklog.Info("Validation for Book upon update", "name", updatingBook.GetName())

	book := &webappv1.Book{}
	if err := v.Client.Get(ctx, client.ObjectKeyFromObject(updatingBook), book); err != nil {
		return admission.Warnings{}, err
	}

	booklog.Info("book", "name", book.Spec.Name)
	booklog.Info("updating book", "name", updatingBook.Spec.Name)

	if updatingBook.Spec.Name == "Python" {
		return admission.Warnings{}, fmt.Errorf("wrong book name %s", updatingBook.Spec.Name)
	}

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type Book.
func (v *BookCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	book, ok := obj.(*webappv1.Book)
	if !ok {
		return nil, fmt.Errorf("expected a Book object but got %T", obj)
	}
	booklog.Info("Validation for Book upon deletion", "name", book.GetName())

	if book.Spec.Name == "Go" {
		return admission.Warnings{}, fmt.Errorf("book name %s can not be deleted, please update to \"\" before delete", book.Spec.Name)
	}

	return nil, nil
}
