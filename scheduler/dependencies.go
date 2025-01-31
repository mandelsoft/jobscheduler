package scheduler

import (
	"github.com/mandelsoft/jobscheduler/scheduler/condition"
)

func DependsOn(jobs ...Job) condition.Condition {
	var list []condition.Condition

	for _, j := range jobs {
		list = append(list, JobFinished(j))
	}
	return condition.And(list...)
}

func DiscardOn(jobs ...Job) condition.Condition {
	var list []condition.Condition

	for _, j := range jobs {
		list = append(list, JobStateReached(j, DISCARDED))
	}
	return condition.Or(list...)
}
