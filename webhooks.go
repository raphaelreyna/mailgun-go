package mailgun

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
)

// GetWebhooks returns the complete set of webhooks configured for your domain.
// Note that a zero-length mapping is not an error.
func (mg *MailgunImpl) GetWebhooks() (map[string]string, error) {
	r := newHTTPRequest(generateDomainApiUrl(mg, webhooksEndpoint))
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	var envelope struct {
		Webhooks map[string]interface{} `json:"webhooks"`
	}
	err := getResponseFromJSON(r, &envelope)
	hooks := make(map[string]string, 0)
	if err != nil {
		return hooks, err
	}
	for k, v := range envelope.Webhooks {
		object := v.(map[string]interface{})
		url := object["url"]
		hooks[k] = url.(string)
	}
	return hooks, nil
}

// CreateWebhook installs a new webhook for your domain.
func (mg *MailgunImpl) CreateWebhook(t string, urls []string) error {
	r := newHTTPRequest(generateDomainApiUrl(mg, webhooksEndpoint))
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	p := newUrlEncodedPayload()
	p.addValue("id", t)
	for _, url := range urls {
		p.addValue("url", url)
	}
	_, err := makePostRequest(r, p)
	return err
}

// DeleteWebhook removes the specified webhook from your domain's configuration.
func (mg *MailgunImpl) DeleteWebhook(t string) error {
	r := newHTTPRequest(generateDomainApiUrl(mg, webhooksEndpoint) + "/" + t)
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	_, err := makeDeleteRequest(r)
	return err
}

// GetWebhookByType retrieves the currently assigned webhook URL associated with the provided type of webhook.
func (mg *MailgunImpl) GetWebhookByType(t string) (string, error) {
	r := newHTTPRequest(generateDomainApiUrl(mg, webhooksEndpoint) + "/" + t)
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	var envelope struct {
		Webhook struct {
			Url string `json:"url"`
		} `json:"webhook"`
	}
	err := getResponseFromJSON(r, &envelope)
	return envelope.Webhook.Url, err
}

// UpdateWebhook replaces one webhook setting for another.
func (mg *MailgunImpl) UpdateWebhook(t string, urls []string) error {
	r := newHTTPRequest(generateDomainApiUrl(mg, webhooksEndpoint) + "/" + t)
	r.setClient(mg.Client())
	r.setBasicAuth(basicAuthUser, mg.APIKey())
	p := newUrlEncodedPayload()
	for _, url := range urls {
		p.addValue("url", url)
	}
	_, err := makePutRequest(r, p)
	return err
}

func (mg *MailgunImpl) VerifyWebhookRequest(req *http.Request) (verified bool, err error) {
	h := hmac.New(sha256.New, []byte(mg.APIKey()))
	io.WriteString(h, req.FormValue("timestamp"))
	io.WriteString(h, req.FormValue("token"))

	calculatedSignature := h.Sum(nil)
	signature, err := hex.DecodeString(req.FormValue("signature"))
	if err != nil {
		return false, err
	}
	if len(calculatedSignature) != len(signature) {
		return false, nil
	}

	return subtle.ConstantTimeCompare(signature, calculatedSignature) == 1, nil
}
