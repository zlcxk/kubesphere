/*

 Copyright 2019 The KubeSphere Authors.

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
package resources

import (
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
	"strings"

	"k8s.io/api/batch/v1beta1"

	"k8s.io/apimachinery/pkg/labels"
)

type cronJobSearcher struct {
}

func (*cronJobSearcher) get(namespace, name string) (interface{}, error) {
	return informers.SharedInformerFactory().Batch().V1beta1().CronJobs().Lister().CronJobs(namespace).Get(name)
}

func cronJobStatus(item *v1beta1.CronJob) string {
	if item.Spec.Suspend != nil && *item.Spec.Suspend {
		return StatusPaused
	}
	return StatusRunning
}

// Exactly Match
func (*cronJobSearcher) match(kv map[string]string, item *v1beta1.CronJob) bool {
	for k, v := range kv {
		switch k {
		case Status:
			if cronJobStatus(item) != v {
				return false
			}
		default:
			if !match(k, v, item.ObjectMeta) {
				return false
			}
		}
	}
	return true
}

func (*cronJobSearcher) fuzzy(kv map[string]string, item *v1beta1.CronJob) bool {
	for k, v := range kv {
		if !fuzzy(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (*cronJobSearcher) compare(a, b *v1beta1.CronJob, orderBy string) bool {
	switch orderBy {
	case LastScheduleTime:
		if a.Status.LastScheduleTime == nil {
			return true
		}
		if b.Status.LastScheduleTime == nil {
			return false
		}
		if a.Status.LastScheduleTime.Equal(b.Status.LastScheduleTime) {
			return strings.Compare(a.Name, b.Name) <= 0
		}
		return a.Status.LastScheduleTime.Before(b.Status.LastScheduleTime)
	default:
		return compare(a.ObjectMeta, b.ObjectMeta, orderBy)
	}
}

func (s *cronJobSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	cronJobs, err := informers.SharedInformerFactory().Batch().V1beta1().CronJobs().Lister().CronJobs(namespace).List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1beta1.CronJob, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = cronJobs
	} else {
		for _, item := range cronJobs {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		return s.compare(result[i], result[j], orderBy)
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}
