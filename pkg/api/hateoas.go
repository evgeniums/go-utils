package api

import (
	"github.com/evgeniums/go-backend-helpers/pkg/access_control"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type HateoasLink struct {
	Target     string `json:"target,omitempty"`
	Operation  string `json:"operation,omitempty"`
	HttpMethod string `json:"http_method,omitempty"`
	Host       string `json:"host,omitempty"`
	Path       string `json:"path,omitempty"`
}

type WithHateoasLinks interface {
	SetHateoasLinks([]*HateoasLink)
}

type ObjectWithHateoasLinks interface {
	common.ID
	SetHateoasLinks([]*HateoasLink)
}

type HateoasLinksStub struct {
}

func (h *HateoasLinksStub) SetHateoasLinks(links []*HateoasLink) {
}

type HateoasLinksContainer struct {
	HateoasLinks []*HateoasLink `json:"_links,omitempty" gorm:"-:all"`
}

func (h *HateoasLinksContainer) SetHateoasLinks(links []*HateoasLink) {
	h.HateoasLinks = links
}

func MakeHateoasLinks(resource Resource, withTestOps ...bool) []*HateoasLink {

	withTest := utils.OptionalArg(false, withTestOps...)
	links := make([]*HateoasLink, 0)

	addLink := func(host string, path string, target string, operation Operation) {

		if operation == nil {
			return
		}
		if operation.TestOnly() && !withTest {
			return
		}

		link := &HateoasLink{}
		link.Host = host
		link.HttpMethod = access_control.Access2HttpMethod(operation.AccessType())
		link.Path = path
		link.Operation = operation.Name()
		link.Target = target

		links = append(links, link)
	}

	// add self operations
	path := resource.FullActualPath()
	selfHost := resource.Host()
	resource.EachOperation(func(operation Operation) error { addLink(selfHost, path, TargetSelf, operation); return nil }, false)

	addGetter := func(target string, r Resource) {
		addLink(r.Host(), r.ActualPath(), target, r.Getter())
	}

	// add children getters
	for _, child := range resource.Children() {
		addGetter(TargetChild, child)
	}

	// add parent getter
	parent := resource.Parent()
	if parent != nil {
		addGetter(TargetParent, parent)
	}

	// done
	return links
}

func InjectHateoasLinksToObject(resource Resource, withLinks WithHateoasLinks) {
	withLinks.SetHateoasLinks(MakeHateoasLinks(resource))
}

func InjectHateoasLinksToList[T ObjectWithHateoasLinks](sampleResource Resource, objects []T) {
	for i := 0; i < len(objects); i++ {
		obj := objects[i]
		sampleResource.SetId(obj.GetID())
		InjectHateoasLinksToObject(sampleResource, obj)
	}
}

func ProcessListResourceHateousLinks[T ObjectWithHateoasLinks](parentResource Resource, resourceType string, objects []T) {
	epResource := parentResource.CloneChain(true)
	sampleResource := NamedResource(resourceType)
	epResource.AddChild(sampleResource)

	InjectHateoasLinksToList(sampleResource, objects)
}
