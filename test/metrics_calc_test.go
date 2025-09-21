package test

import (
	"testing"
	"time"

	"admira/internal/metrics"
	"admira/internal/model"
)

func ymd(s string) time.Time { t,_ := time.Parse("2006-01-02", s); return t }

func TestChannelMetricsBasic(t *testing.T) {
	ads := []model.AdsPerf{{
		Date: ymd("2025-08-01"), CampaignID: "C-1001", Channel: "google_ads",
		Clicks: 100, Impressions: 1000, Cost: 200,
		UTMCampaign: "back_to_school", UTMSource: "google", UTMMedium: "cpc",
	}}
	crm := []model.CRMOpp{
		{OpportunityID:"O-1", Stage:"lead", CreatedAt: ymd("2025-08-01"), UTMCampaign:"back_to_school", UTMSource:"google", UTMMedium:"cpc"},
		{OpportunityID:"O-2", Stage:"opportunity", CreatedAt: ymd("2025-08-01"), UTMCampaign:"back_to_school", UTMSource:"google", UTMMedium:"cpc"},
		{OpportunityID:"O-3", Stage:"closed_won", Amount: 5000, CreatedAt: ymd("2025-08-01"), UTMCampaign:"back_to_school", UTMSource:"google", UTMMedium:"cpc"},
	}
	rows := metrics.ChannelMetrics(ads, crm, metrics.Filters{})
	if len(rows) != 1 { t.Fatalf("expected 1 row") }
	r := rows[0]
	if r.CPC == 0 || r.ROAS == 0 { t.Fatalf("expected non-zero CPC/ROAS") }
	if r.ClosedWon != 1 || r.Revenue != 5000 { t.Fatalf("won/revenue mismatch") }
}
