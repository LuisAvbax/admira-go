package model

import "time"

type AdsPerf struct {
	Date        time.Time
	CampaignID  string
	Channel     string
	Clicks      int
	Impressions int
	Cost        float64
	UTMCampaign string
	UTMSource   string
	UTMMedium   string
}

type CRMOpp struct {
	OpportunityID string
	ContactEmail  string
	Stage         string // lead, opportunity, closed_won, ...
	Amount        float64
	CreatedAt     time.Time
	UTMCampaign   string
	UTMSource     string
	UTMMedium     string
}

// Claves de idempotencia:
func (a AdsPerf) Key() string  { return a.Date.Format("2006-01-02")+"|"+a.CampaignID+"|"+a.Channel }
func (o CRMOpp) Key() string   { return o.OpportunityID }
