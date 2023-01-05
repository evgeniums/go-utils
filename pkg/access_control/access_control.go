package access_control

import (
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
)

// Interface for access controllers.
type AccessControl interface {
	CheckAccess(ctx op_context.Context, resource Resource, subject Subject, accessType AccessType) bool
	DefaultAccess() Access
	SetDefaultAccess(Access)
}

// Base implementation of access controller.
type AccessControlBase struct {
	acl             Acl
	resourceManager ResourceManager
	defaultAccess   Access
}

func (a *AccessControlBase) CheckAccess(ctx op_context.Context, resource Resource, subject Subject, accessType AccessType) (bool, error) {

	ctx.TraceInMethod("AccessControl.CheckAccess")

	// use default access
	allow := a.defaultAccess.Check(accessType)
	ruleFound := false

	// collect tags of the resource using tags of all resources in the path
	tags := make(map[int][]string, 0)
	paths := resource.Paths()
	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]
		resourceTags, err := a.resourceManager.ResourceTags(ctx, path)
		if err != nil {
			ctx.TraceOutMethod()
			return false, err
		}
		tags[i] = resourceTags
	}

	// look for rules for all resource paths starting from the most specific path
	for i := len(paths) - 1; i >= 0; i-- {
		path := paths[i]
		// look for rule per each tag starting from the tags for the most specific path
		for j := i; j >= 0; j-- {
			resourceTags := tags[j]
			// if rule for tag is not found then look for the resource without tags
			resourceTags = append(resourceTags, "")
			for _, tag := range resourceTags {
				// look for the first matching role a rule esists for
				for _, role := range subject.Roles() {
					rule, err := a.acl.FindRule(ctx, path, tag, role)
					if err != nil {
						ctx.TraceOutMethod()
						return false, err
					}
					if rule != nil {
						ruleFound = true
						allow = rule.Access().Check(accessType)
						// TODO log used rule
						break
					}
				}
				if ruleFound {
					break
				}
			}
			if ruleFound {
				break
			}
		}
		if ruleFound {
			break
		}
	}

	// TODO log detected access
	ctx.TraceOutMethod()
	return allow, nil
}

func (a *AccessControlBase) DefaultAccess() Access {
	return a.defaultAccess
}

func (a *AccessControlBase) SetDefaultAccess(access Access) {
	a.defaultAccess = access
}
