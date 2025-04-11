package jobnet

import (
	"github.com/mandelsoft/jobscheduler/scheduler"
)

func DependsOn(jobs ...string) Condition {
	var list []Condition

	for _, j := range jobs {
		list = append(list, JobFinished(j))
	}
	return And(list...)
}

func DiscardOn(jobs ...string) Condition {
	var list []Condition

	for _, j := range jobs {
		list = append(list, JobStateReached(j, scheduler.DISCARDED))
	}
	return Or(list...)
}
