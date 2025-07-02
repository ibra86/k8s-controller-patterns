package api

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"context"

	frontendv1alpha1 "github.com/ibra86/k8s-controller-patterns/pkg/apis/frontend/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type FrontendPageAPI struct {
	K8sClient client.Client
	Namespace string
}

// @Description FrontendPage resource (Swagger only)
type FrontendPageDoc struct {
	Name     string `json:"name" example:"example-page"`
	Contents string `json:"contents" example:"<h1>Hello</h1>"`
	Image    string `json:"image" example:"nginx:latest"`
	Replicas int    `json:"replicas" example:"2"`
}

// @Description List of FrontendPage resources (Swagger only)
type FrontendPageListDoc struct {
	Items []FrontendPageDoc `json:"items"`
}

// ListFrontendPage godoc
// @Summary List all FrontendPages
// @Description Get all FrontendPage resources
// @Tags frontendpages
// @Produce json
// @Success 200 {object} FrontendPageListDoc
// @Router /api/frontendpages [get]
func (api *FrontendPageAPI) ListFrontendPages(ctx *fasthttp.RequestCtx) {
	list := &frontendv1alpha1.FrontendPageList{}
	err := api.K8sClient.List(context.Background(), list, client.InNamespace(api.Namespace))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(fmt.Sprintf(`{"error":"%v"}`, err))
		return
	}
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(list.Items); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "failed to encode JSON"}`)
		return
	}
}

// GetFrontendPage godoc
// @Summary Get a FrontendPage
// @Description Get a FrontendPage by name
// @Tags frontendpages
// @Produce json
// @Param name path string true "FrontendPage name"
// @Success 200 {object} FrontendPageDoc
// @Failure 404 {object} map[string]string
// @Router /api/frontendpages{name} [get]
func (api *FrontendPageAPI) GetFrontendPage(ctx *fasthttp.RequestCtx) {
	nameVal := ctx.UserValue("name")
	if nameVal == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error": "missing name parameter"}`)
		return
	}
	name := nameVal.(string) // try to extract the underlying value as a string.
	obj := &frontendv1alpha1.FrontendPage{}
	err := api.K8sClient.Get(
		context.Background(),
		client.ObjectKey{Namespace: api.Namespace, Name: name},
		obj,
	)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString(fmt.Sprintf(`{"error": "%v"}`, err))
		return
	}
	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(obj); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "failed to encode JSON"}`)
		return
	}
}

// CreateFrontendPage godoc
// @Summary Create a FrontendPage
// @Description Create a FrontendPage
// @Tags frontendpages
// @Accept json
// @Produce json
// @Param body body FrontendPageDoc true "FrontendPage object"
// @Success 201 {object} FrontendPageDoc
// @Failure 404 {object} map[string]string
// @Router /api/frontendpages [post]
func (api *FrontendPageAPI) CreateFrontendPage(ctx *fasthttp.RequestCtx) {
	obj := &frontendv1alpha1.FrontendPage{}
	if err := json.Unmarshal(ctx.PostBody(), obj); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(fmt.Sprintf(`{"error": "%v"}`, err))
		return
	}

	obj.Namespace = api.Namespace
	if err := api.K8sClient.Create(context.Background(), obj); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(fmt.Sprintf(`{"error": "%v"}`, err))
		return
	}
	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(obj)
	if err := json.NewEncoder(ctx).Encode(obj); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "failed to encode JSON"}`)
		return
	}
}

// UpdateFrontendPage godoc
// @Summary Update a FrontendPage
// @Description Update a FrontendPage
// @Tags frontendpages
// @Accept json
// @Produce json
// @Param name path string true "FrontendPage name"
// @Param body body FrontendPageDoc true "FrontendPage object"
// @Success 200 {object} FrontendPageDoc
// @Failure 400 {object} map[string]string
// @Router /api/frontendpages/{name} [put]
func (api *FrontendPageAPI) UpdateFrontendPage(ctx *fasthttp.RequestCtx) {
	nameVal := ctx.UserValue("name")
	if nameVal == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error": "missing name parameter"}`)
		return
	}
	name := nameVal.(string) // try to extract the underlying value as a string.
	existing := &frontendv1alpha1.FrontendPage{}
	err := api.K8sClient.Get(
		context.Background(),
		client.ObjectKey{Namespace: api.Namespace, Name: name},
		existing,
	)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString(fmt.Sprintf(`{"error": "%v"}`, err))
		return
	}

	var patch struct {
		Spec frontendv1alpha1.FrontendPageSpec `json:"spec"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &patch); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(fmt.Sprintf(`{"error": "%v"}`, err))
		return
	}
	existing.Spec = patch.Spec

	if err := api.K8sClient.Update(context.Background(), existing); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(fmt.Sprintf(`{"error": "%v"}`, err))
		return
	}

	ctx.SetContentType("application/json")
	if err := json.NewEncoder(ctx).Encode(existing); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBodyString(`{"error": "failed to encode JSON"}`)
		return
	}
}

// DeleteFrontendPage godoc
// @Summary Delete a FrontendPage
// @Description Delete a FrontendPage
// @Tags frontendpages
// @Param name path string true "FrontendPage name"
// @Success 204 {object} nil
// @Failure 404 {object} map[string]string
// @Router /api/frontendpages/{name} [delete]
func (api *FrontendPageAPI) DeleteFrontendPage(ctx *fasthttp.RequestCtx) {
	nameVal := ctx.UserValue("name")
	if nameVal == nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString(`{"error": "missing name parameter"}`)
		return
	}
	name := nameVal.(string) // try to extract the underlying value as a string.
	obj := &frontendv1alpha1.FrontendPage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: api.Namespace,
		},
	}

	ctx.SetContentType("application/json")
	if err := api.K8sClient.Delete(context.Background(), obj); err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.SetBodyString(fmt.Sprintf(`{"error": "%v"}`, err))
		return
	}
}
