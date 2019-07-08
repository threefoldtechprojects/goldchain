package main

import (
	"fmt"
	"html/template"

	"github.com/nbh-digital/goldchain/pkg/config"
)

func mustTemplate(title, text string) *template.Template {
	p := template.New(title)
	return template.Must(p.Parse(text))
}

// RequestBody is used to render the request.html template
type RequestBody struct {
	ChainName    string
	ChainNetwork string
	CoinUnit     string
	Error        string
}

var requestTemplate = mustTemplate("request.html", fmt.Sprintf(`
<head>
	<title>{{.CoinUnit}} Faucet</title>
</head>
<body>
	<div align="center">
		<h1 style="margin-top:3em">{{.ChainName}} {{.ChainNetwork}} faucet</h1>

		{{if .Error}}
		<div style="margin:50px;display:inline-flex;align-items:center;border:3px solid red;padding:10px;background:#ffe5e5;">
			<div style="font-size:80px;border:2px solid red;border-radius:50%%;width:80px;color:red;line-height:80px;">!</div>
			<div style="color:red;display:inline-block;padding: 0 20px;font-weight:bold">{{.Error}}</div>
		</div>
		{{end}}

		<h3>Request %[1]d {{.CoinUnit}} by entering your address below and submitting the form.</h3>
		<form action="/request/tokens" method="POST">
			<div>Address: <input type="text" size="78" name="uh"></div>
			<br>
			<div><input type="submit" value="Request %[1]d {{.CoinUnit}}" style="width:20em;height:2em;font-weight:bold;font-size:1em;"></div>
		</form>

		<h3 style="margin-top:50px;">Request authorization or deauthorization by entering your address below and submitting the form.</h3>
		<form action="/request/authorize" method="POST">
			<div>Address: <input type="text" size="78" name="uh"></div>
			<br>
			<input type="radio" name="authorize" value="true" checked>Authorize<br>
			<input type="radio" name="authorize" value="false">Deauthorize<br>
			<br>
			<div><input type="submit" value="Request address authorization update" style="width:20em;height:2em;font-weight:bold;font-size:1em;"></div>
		</form>
	
		<div style="margin-top:50px;"><small>{{.ChainName}} faucet v%s</small></div>
	</div>
</body>
`, coinsToGive, config.Version.String()))

// CoinConfirmationBody is used to render the coinconfirmation.html template
type CoinConfirmationBody struct {
	ChainName     string
	ChainNetwork  string
	CoinUnit      string
	Address       string
	TransactionID string
}

var coinConfirmationTemplate = mustTemplate("coinconfirmation.html", fmt.Sprintf(`
<head>
	<title>{{.CoinUnit}} Faucet</title>
</head>
<body>
	<div align="center">
		<h1>%d {{.CoinUnit}} succesfully transferred on {{.ChainName}}'s {{.ChainNetwork}} to {{.Address}}</h1>
		<p>You can look up the transaction using the following ID:</p>
		<div><code>{{.TransactionID}}</code></div>
		<div style="margin-top:50px;"><small>{{.ChainName}} faucet v%s</small></div>
	</div>
</body>
`, coinsToGive, config.Version.String()))

// AuthorizationConfirmationBody is used to render the authorizationconfirmation.html page
type AuthorizationConfirmationBody struct {
	ChainName     string
	ChainNetwork  string
	CoinUnit      string
	Address       string
	Action        string
	TransactionID string
}

var authorizationConfirmationTemplate = mustTemplate("authorizationconfirmation.html", fmt.Sprintf(`
<head>
	<title>{{.CoinUnit}} Faucet</title>
</head>
<body>
	<div align="center">
		<h1>Succesfully {{.Action}} address {{.Address}} on {{.ChainName}}'s {{.ChainNetwork}}</h1>
		<p>You can look up the transaction using the following ID:</p>
		<div><code>{{.TransactionID}}</code></div>
		<div style="margin-top:50px;"><small>{{.ChainName}} faucet v%s</small></div>
	</div>
</body>
`, config.Version.String()))
