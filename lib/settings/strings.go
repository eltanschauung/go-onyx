package settings

import (
	"git.gammaspectra.live/git/go-away/utils"
)

var DefaultStrings = utils.NewStrings(map[string]string{
	"title_challenge": "Confirming you are a whale",
	"title_error":     "Oh no!",

	"noscript_warning": "<p>Sadly, you may need to enable JavaScript to get past this challenge. This is required because AI companies have changed the social contract around how website hosting works.</p>",

	"details_title": "Why am I seeing this?",
	"details_text": `
<p>
	To avoid this Proof of Work (PoW) check, try entering a password at <a href="/auth">bantculture.com/auth</a>.
</p>
`,
	"details_contact_admin_with_request_id": "If you have any issues contact the site administrator and provide the following Request Id",

	"button_refresh_page": "Refresh page",

	"status_loading_challenge":   "Loading challenge",
	"status_starting_challenge":  "Starting challenge",
	"status_loading":             "Loading...",
	"status_calculating":         "Onyx is pouring sproke into your computer...",
	"status_challenge_success":   "Challenge success!",
	"status_challenge_done_took": "Done! Took",
	"status_error":               "Error:",
})
