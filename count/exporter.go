package count

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

type ClusterManager struct {
	AddCodeCountDesc *prometheus.Desc
	DelCodeUsageDesc *prometheus.Desc
	GitRepo          string
	Branch           string
	User             string
}

func (c *ClusterManager) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.AddCodeCountDesc
	ch <- c.DelCodeUsageDesc
}

func (c *ClusterManager) Collect(ch chan<- prometheus.Metric) {
	for gitRepo, v := range CountData {
		for branch, d := range v {
			for user, data := range d {
				add := data["add"]
				shortGitRepo := strings.Replace(gitRepo, "git@github.com:kwen94/", "", -1)
				shortBranch := strings.Replace(branch, "remotes/origin/", "", -1)
				ch <- prometheus.MustNewConstMetric(
					c.AddCodeCountDesc,
					prometheus.CounterValue,
					float64(add),
					shortGitRepo,
					shortBranch,
					user,
				)
				del := data["del"]
				ch <- prometheus.MustNewConstMetric(
					c.DelCodeUsageDesc,
					prometheus.CounterValue,
					float64(del),
					shortGitRepo,
					shortBranch,
					user,
				)
			}
		}
	}

}

func NewClusterManager() *ClusterManager {
	return &ClusterManager{
		//Zone: zone,
		AddCodeCountDesc: prometheus.NewDesc(
			"add_code_count_total",
			"Add Code Count Total.",
			[]string{"gitRepo", "branch", "user"},
			//prometheus.Labels{"zone": zone},
			nil,
		),
		DelCodeUsageDesc: prometheus.NewDesc(
			"del_code_count_total",
			"Del Code Count Total.",
			[]string{"gitRepo", "branch", "user"},
			//prometheus.Labels{"zone": zone},
			nil,
		),
	}
}
