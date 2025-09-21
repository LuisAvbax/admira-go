package ingest

import (
	"context"
	"strings"
	"time"

	"admira/internal/model"
	"admira/internal/store"
	"admira/internal/util"
)

type AdsEnvelope struct {
	External struct {
		Ads struct {
			Performance []struct {
				Date        string  `json:"date"`
				CampaignID  string  `json:"campaign_id"`
				Channel     string  `json:"channel"`
				Clicks      int     `json:"clicks"`
				Impressions int     `json:"impressions"`
				Cost        float64 `json:"cost"`
				UTMCampaign string  `json:"utm_campaign"`
				UTMSource   string  `json:"utm_source"`
				UTMMedium   string  `json:"utm_medium"`
			} `json:"performance"`
		} `json:"ads"`
	} `json:"external"`
}
type CRMEnvelope struct {
	External struct {
		CRM struct {
			Opportunities []struct {
				OpportunityID string  `json:"opportunity_id"`
				ContactEmail  string  `json:"contact_email"`
				Stage         string  `json:"stage"`
				Amount        float64 `json:"amount"`
				CreatedAt     string  `json:"created_at"`
				UTMCampaign   string  `json:"utm_campaign"`
				UTMSource     string  `json:"utm_source"`
				UTMMedium     string  `json:"utm_medium"`
			} `json:"opportunities"`
		} `json:"crm"`
	} `json:"external"`
}

type ETL struct {
	ADSURL string
	CRMURL string
	HTTP   *HTTPClient
	Store  *store.Memory
}

func (e *ETL) Run(ctx context.Context, since *time.Time) error {
	// 1) Fetch
	var ads AdsEnvelope
	var crm CRMEnvelope
	if err := e.HTTP.GetJSON(ctx, e.ADSURL, &ads); err != nil { return err }
	if err := e.HTTP.GetJSON(ctx, e.CRMURL, &crm); err != nil { return err }

	// 2) Normalize + filter + upsert
	for _, r := range ads.External.Ads.Performance {
		dt := util.ParseYMD(r.Date)
		if since != nil && dt.Before(*since) { continue }
		a := model.AdsPerf{
			Date:        dt,
			CampaignID:  strings.TrimSpace(r.CampaignID),
			Channel:     strings.ToLower(strings.TrimSpace(r.Channel)),
			Clicks:      util.NonNegInt(r.Clicks),
			Impressions: util.NonNegInt(r.Impressions),
			Cost:        util.NonNegFloat(r.Cost),
			UTMCampaign: strings.ToLower(strings.TrimSpace(r.UTMCampaign)),
			UTMSource:   strings.ToLower(strings.TrimSpace(r.UTMSource)),
			UTMMedium:   strings.ToLower(strings.TrimSpace(r.UTMMedium)),
		}
		e.Store.UpsertAds(a) // idempotente por clave
	}

	for _, r := range crm.External.CRM.Opportunities {
		created := util.ParseRFC3339(r.CreatedAt)
		if since != nil && created.Before(*since) { continue }
		o := model.CRMOpp{
			OpportunityID: strings.TrimSpace(r.OpportunityID),
			ContactEmail:  strings.TrimSpace(r.ContactEmail),
			Stage:         strings.ToLower(strings.TrimSpace(r.Stage)),
			Amount:        util.NonNegFloat(r.Amount),
			CreatedAt:     created,
			UTMCampaign:   strings.ToLower(strings.TrimSpace(r.UTMCampaign)),
			UTMSource:     strings.ToLower(strings.TrimSpace(r.UTMSource)),
			UTMMedium:     strings.ToLower(strings.TrimSpace(r.UTMMedium)),
		}
		e.Store.UpsertCRM(o)
	}
	return nil
}
