package components

import (
	"encoding/json"

	"github.com/a-h/templ"

	"github.com/Ankumeah/JSBEE/backend/internal/provider"
)

// jsonLD renders the site-wide JSON-LD structured data (Organization +
// WebSite). Built in Go rather than inline in base.templ because templ does
// not evaluate expressions inside <script type="application/ld+json"> blocks.
func jsonLD() templ.Component {
	site := provider.SiteURL
	doc := map[string]any{
		"@context": "https://schema.org",
		"@graph": []any{
			map[string]any{
				"@type":         "Organization",
				"name":          "JSBEE",
				"alternateName": "Journal of Sustainable Business in Emerging Economies",
				"url":           site + "/",
				"logo":          site + "/favicon.png",
				"sameAs": []string{
					"https://www.instagram.com/jsbee.research",
					"https://www.linkedin.com/company/journal-of-sustainable-business-in-emerging-economies",
				},
			},
			map[string]any{
				"@type": "WebSite",
				"name":  "JSBEE",
				"url":   site + "/",
			},
		},
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return templ.Raw("")
	}
	return templ.Raw(`<script type="application/ld+json">` + string(raw) + `</script>`)
}
