package metrics

import (
	"strings"
	"time"

	"admira/internal/model"
	"admira/internal/util"
)

type Filters struct {
	From, To           *time.Time
	Channel, Campaign  string
	UTMCampaign        string
	UTMSource          string
	UTMMedium          string
	Limit, Offset      int
}

type ChannelRow struct {
	Date         string  `json:"date"`
	Channel      string  `json:"channel"`
	CampaignID   string  `json:"campaign_id"`
	Clicks       int     `json:"clicks"`
	Impressions  int     `json:"impressions"`
	Cost         float64 `json:"cost"`
	Leads        int     `json:"leads"`
	Opportunities int    `json:"opportunities"`
	ClosedWon    int     `json:"closed_won"`
	Revenue      float64 `json:"revenue"`
	CPC          float64 `json:"cpc"`
	CPA          float64 `json:"cpa"`
	CVRLeadToOpp float64 `json:"cvr_lead_to_opp"`
	CVROppToWon  float64 `json:"cvr_opp_to_won"`
	ROAS         float64 `json:"roas"`
}

func inRange(t time.Time, from, to *time.Time) bool {
	if from != nil && t.Before(*from) { return false }
	if to != nil   && t.After(*to)    { return false }
	return true
}
func like(a, b string) bool { return b == "" || strings.EqualFold(a, b) }

func ChannelMetrics(ads []model.AdsPerf, crm []model.CRMOpp, f Filters) []ChannelRow {
	// Índices CRM por UTM triple y por fecha
	type key struct{ uc, us, um string }
	crmByUTM := map[key][]model.CRMOpp{}
	for _, o := range crm {
		if !inRange(o.CreatedAt, f.From, f.To) { continue }
		k := key{strings.ToLower(o.UTMCampaign), strings.ToLower(o.UTMSource), strings.ToLower(o.UTMMedium)}
		crmByUTM[k] = append(crmByUTM[k], o)
	}

	// Agregación por (fecha, canal, campaign)
	type gk struct{ d, ch, cid string }
	acc := map[gk]*ChannelRow{}

	for _, a := range ads {
		if !inRange(a.Date, f.From, f.To) { continue }
		if !like(a.Channel, f.Channel) || !like(a.CampaignID, f.Campaign) { continue }
		if !like(a.UTMCampaign, f.UTMCampaign) || !like(a.UTMSource, f.UTMSource) || !like(a.UTMMedium, f.UTMMedium) { continue }

		g := gk{d: a.Date.Format("2006-01-02"), ch: a.Channel, cid: a.CampaignID}
		row := acc[g]
		if row == nil {
			row = &ChannelRow{Date: g.d, Channel: g.ch, CampaignID: g.cid}
			acc[g] = row
		}
		row.Clicks += a.Clicks
		row.Impressions += a.Impressions
		row.Cost += a.Cost

		// CRM join por UTM (fallback si es necesario)
		matches := crmByUTM[key{a.UTMCampaign, a.UTMSource, a.UTMMedium}]
		if len(matches) == 0 && a.UTMCampaign != "" {
			// fallback por campaign only
			for k2, lst := range crmByUTM {
				if k2.uc == a.UTMCampaign {
					matches = append(matches, lst...)
				}
			}
		}
		for _, o := range matches {
			if !inRange(o.CreatedAt, f.From, f.To) { continue }
			switch o.Stage {
			case "lead":
				row.Leads++
			case "opportunity":
				row.Opportunities++
			case "closed_won":
				row.ClosedWon++
				row.Revenue += o.Amount
			}
		}
	}

	// Derivadas
	out := make([]ChannelRow, 0, len(acc))
	for _, r := range acc {
		r.CPC = util.SafeDiv(r.Cost, float64(r.Clicks))
		r.CPA = util.SafeDiv(r.Cost, float64(r.Leads))
		r.CVRLeadToOpp = util.SafeDiv(float64(r.Opportunities), float64(r.Leads))
		r.CVROppToWon  = util.SafeDiv(float64(r.ClosedWon), float64(r.Opportunities))
		r.ROAS = util.SafeDiv(r.Revenue, r.Cost)
		out = append(out, *r)
	}

	// Paginación simple
	start := f.Offset
	if start > len(out) { return []ChannelRow{} }
	end := start + f.Limit
	if f.Limit <= 0 || end > len(out) { end = len(out) }
	return out[start:end]
}
