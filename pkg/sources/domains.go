package sources

import (
	"github.com/ip-rw/rank/pkg/util"
	"github.com/sirupsen/logrus"
)

type DomainSource interface {
	Name() string
	Lookup(string) ([]string, error)
}

func FindPossibleDomains(c *util.Company) []string {
	company := c.Name
	//pc := c.RegisteredAddress.PostalCode
	var urls []string
	var modules = map[DomainSource]string {
		TLD{}: company,
		DuckDuckGo{}: company + " \"" + c.CompanyNumber + "\"",
		Clearbit{}: company	,
	}

	for m, search := range modules {
		if res, err := m.Lookup(search); err != nil {
			logrus.WithError(err).WithField("source", m.Name()).Error("source error")
		} else {
			for _, domain := range res {
				logrus.WithField("source", m.Name()).WithField("domain", domain).WithField("search", CleanCompanyName(search)).Debug("new domain")
				urls = util.AppendUniq(urls, domain)
			}
		}
	}
	return urls
}
