// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

// +build !recording

package helpers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"

	bm "github.com/oracle/bmcs-go-sdk"
)

const RUNMODE = RunmodeReplay

const (
	mockKey = `
-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,EB2F52157F4A19A2

iWOPMqtbRxstDZDCKeQsOWFgZ3XdMCfz84MWmB5imkM6D34kOj8BXQFIcc8Hi2Lj
YQQ03HcpSE8h8Fj5086frRDEl5Sd1QmThiT6DFsbvNom9coat1zwUZHoMIgY+cIH
LuIrlAVQaYidVC5MBgFoJlitszQpZnygnN8H8MMQS9WIIeV5wVqhpVZMzjp9Unq0
RfSeryHIIH0ZwgSelnMbec5SRiMaWtqcjpL9p0qDIwmxdJW2JY3sDNQ2S+V1wNG4
kXw8Bfd3+F7pw+s3xmZVVmdfNCHzWhGSV2ssOO9yovAd0tark2FNTYZTgx4HzijY
i11cAqgB8wfOpEe7RXleKhjJRdtto1ANeAx3wfTW0As7zXdoZCJy/7qXBfVBe+nA
I6j5fGba3b2kkDnrd740d4Na7/fHnMKt+j+Ke4+C/Kimxt6vuQtvb9brOjV+qpPW
6yLzyXaKtoZaGFi5839F08fr/2yY8huz7H5C1U6Gza2/8ZvIRhvCKJOvRYrTnnGT
08QsX+wzYNqUyl5j64L8QPACY3zdWpuxQ7FhM2fMnymDKJjBhoRbnEjxJppaES/Q
tPo6w7nB4sMHhOobyXG+12kToEl/BVnoL/KG77M4kTCno6tg4Ca8+cvw2y/dFj3u
KD+Tt3j4uITc4Majt9oNyWHINlM6lV0czF852RliwdZgf+yQxi734QeJj8XhcDNK
Pt2fiFxsVuy3JA4CYcfn/DuKGqrMwHE2rn1e9tQ8nzsZn84k+oLBHC4LBoEtCg/r
BxqK4FQpUzccw5hXGwvZobE6qnp77m0mkgukGmo0ieyumhlQt3lNGq8P4WQDFF0b
KnJzEf7Z58olBmkn8qtAGuYTEN9/1/LXMXE9RKTMe/5N7zJFqSIzSFasP58PMLmP
U+/dGRkXxMSXO8LvQm5JpY1K5fOjeZC5ePSmHlx5i9cxRJfU05W7XrNAdCwn4F3I
tji5wcCKFziOI2JREZSR5SjfxEbhUbdYQ/eEkzPSVfvPCBCBAxwMyy4NipRX0e3v
jeVBnR1Oxz19i85yOwHpsYARKcMY30o+aSpNafDp8jSKx6dJTTOKdaZ1X/IpNTdt
ppMgoicYof0es7rpsd9rB+n8l5LsspbTKq9f3JejqZC9FW8G/ZakmTuMPSc+ITub
zaiiRjN6+CJCA2t8l9zTFLrVb71H0/e9yZAW/TxFVWnkLJqZkFn1LGZxkQUOARQW
+/UDhzb/9rSypG+EHJD9MQDa//S6YnzdybpOq9dHg2Ttf1sevb8aLLRa2AL7s8cs
q8RuCTx7VKIkskVKcx1W8Cf56bZn4mZ95fzlNMbyIJNw+ZFcWKX1L54EpBI9i7zb
vUSXT2wUYYPMAF+Zbksa0aTU9rkqRP+gKmYaPDcO5kchegHjOfI5oqzJcZ+3udbV
sx3gkL1HT1nnFxKpzDeroE1YenOHF8WCbEUi6P1lDh7L6MhFfdFf5oIXL35pvui7
GKOEzfUTeaEajeTogLYl/StW6i8+07acJ7Edk//jvqATf5jIG7hVHNQyOpwM8W/K
cMuGz30TqwTxIXNqqtKzz94jAigz3YWtJbWgBuphTqZ9xTWhPeipRw==
-----END RSA PRIVATE KEY-----
`

	tenancyOcid = "ocid1.tenancy.oc1..aaaaaaaayfzsknabeptheebqsaicjddtlubq7dnwz5izbvs3vfs4xmkargta"
	userOcid    = "ocid1.user.oc1..aaaaaaaaubrcg3qvtqf2ormvbgjxty5bpj5acpjw2q7ul4pb3pha2vf56tqa"
	fingerprint = "3a:09:91:bb:d1:66:a0:13:12:f1:bf:3e:f6:ea:16:09"
)

var re = regexp.MustCompile(`(.*compartmentId=)ocid[a-zA-Z0-9.%]+(&?.*)`)

func GetClient(cassetteName string) *TestClient {
	rec, err := recorder.NewAsMode(cassetteName, recorder.ModeReplaying, nil)

	// Since the default matcher only matches by url, we inject a header with a sequential request number
	// so we can find the appropriate request when we replay
	rec.SetMatcher(func(r *http.Request, cr cassette.Request) bool {

		// maybe it's an exact match, return early
		if cassette.DefaultMatcher(r, cr) && (r.Header.Get("Request-Number") == cr.Headers.Get("Request-Number")) {
			return true
		}

		// matching by url might not work if the user who recorded the test data had a different config from
		// the user who is running the tests. Stripping compartmentId from the url accommodates that.
		// Tripping hazard: theoretical cross-compartment acceptance could fail to replay due to this but
		// most likely not due to the additional Request-Number provision.
		subStr := re.ReplaceAllLiteralString(r.URL.String(), `$1$2`)
		recStr := re.ReplaceAllLiteralString(cr.URL, `$1$2`)
		return (recStr == subStr) && (r.Header.Get("Request-Number") == cr.Headers.Get("Request-Number"))
	})

	if err != nil {
		panic(fmt.Sprintf("could not create recorder. error: %s", err))
	}

	client, err := bm.NewClient(
		userOcid,
		tenancyOcid,
		fingerprint,
		bm.PrivateKeyBytes([]byte(mockKey)),
		bm.PrivateKeyPassword("supersecret"),
		bm.CustomTransport(NewTestTransport(rec)),
	)

	if err != nil {
		panic(fmt.Sprintf("could not create client. error: %s", err))

	}

	tc := &TestClient{
		Client:   client,
		recorder: rec,
	}

	return tc
}
