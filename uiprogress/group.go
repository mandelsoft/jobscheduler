package uiprogress

import (
	"github.com/mandelsoft/jobscheduler/uiprogress/ppi"
)

type Group interface {
	Container

	SetFollowUpGap(gap string) Group
	Gap() string

	ppi.ProgressInterface[Group]
	SetSpeed(int) Group
}

type _group struct {
	*ppi.GroupBase[Group, Spinner]
	main Spinner
}

var _ Group = (*_group)(nil)

func NewGroup(p Container, gap string, set int) Group {
	g := &_group{}
	g.GroupBase, g.main = ppi.NewGroupBase[Group, Spinner](p, g, gap, func(b *ppi.GroupBase[Group, Spinner]) (Spinner, bool) {
		s := NewSpinner(b, set)
		return s, s != nil // cannot check for nil by caller because of parametric type
	})
	return g
}

func (g *_group) SetPhases(p ...string) Group {
	g.main.SetPhases(p...)
	return g
}

func (g *_group) SetSpeed(i int) Group {
	g.main.SetSpeed(i)
	return g
}
