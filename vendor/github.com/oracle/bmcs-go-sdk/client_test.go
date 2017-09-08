// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package baremetal

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testPrivateKey = []byte(
	`-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,9F4D00DEF02B2B75

IbSQEhNjPeRt49jUhZbhAEaAIG4L9IokDksw/P/QdCPXzZT008xzYK/zmxkz7so1
ZwvIYHn07E0Ul6fIHR6kjw/+MD7AWluCN1FLHs3PHc4XF4THUCKFCC90FvGJ2PEs
kEh7oJ4azZA/PH51g4rSgWpYtH5B/S6ioE2eZ9jJ/prH+34pCuOpX4AvXEFl5zue
pjFm5FhsReAhZ/9eCvjgjIWDHKc7PRfinwSydVHQSzgDnuq+GTMzQh6eztS+EuAp
MLg7w0mazTqmPOuMT+mw9SHGaIePGzA9TcwB1y3QgkYsg3Ch20uN/sUymgQ4PEKI
njXLldWDYvFvv1Tv3/8IOjCEodQ4P/5oWz7msrLh3QF+EhF7lQPYO7132e9Hvz3C
hTmcygmVGrPCtOY1jzuqy+/Kmt4Gv8FQpSnO7i8wFvt5v0N26av18RO10CzYY1ut
EV6WvynimFUtg1Lo03cadh7bspNohSXfFLpbNTji5NwHrIa+UQqTw3h4/zSPZHJl
NwHwM2I8N5lcCsqmSbM01+uTRG3QZ5i1BS8fsArHaAcvPyLvOy4mZGKkpuNlLDXo
qrCCsb+0m9jHR2bzx5AGp4impdHm2Qi3vTV3dMe277wqKkU5qfd5yDbL2eTqAYzQ
hXpPmTjquOTNYdbvoNsOg4TCHZv7WCsGY0nNMPrRO7zXCDApA6cKDJzagbqhW5Zu
/yz7sDT2D3wzE2WXUbtIBLevXyF0OS3AL7AgfbcyAviByOfmEb7WCP9jmdCFaLwY
SgNh9AjeOgkEEr/cRg1kBAXt0kuE7By0w+/ODJHZYelG0wg5nxhseA9Kc596XIJl
NyjbL87CXGfXmMoSYYTA4rzbtCDMmee7xHtbWiYKF1VGxNaGkQ5nnZSJLhCaI6rH
AD0XYwxv92j4fIjHqonbY/dlIKPot1t3VRcdnebbZMjAcNZ63n+I/iVla3DJpWLO
1gT50A4H2uEAve+WWFWmDQe2rfg5wwUtVVkot+Tn3McB6RzNqgcs0c+7uNDnDcOB
WtQ1OfniE1TdoFCPfYcDw8ngimw7uMYwp4mZIYtwlk7Z5GFl4YpNQeLOgh368ao4
8HL7EnTZmiU5cMbuaA8cZmUbgBqiQY0DtLF22VquThi0QOeUMJxJ6N1QUPckD3AU
dikEn0gilOsDQ51fnOsgk9J2uCz8rd5bnyUXlIguj5pyz6S7agyYFhRrXessVzHd
3889QM9V82+px5mv4qCvMn6ReYOvC+KSY1hn4ljXsndOM+6hQzD5CZKeL948pXRn
G7nqbG9D44wLklOz6mkIvqLn3qxEFWapl9UK7yfzjoezGoqeNFweadZ10Kp2+Umu
Sa759/2YDCZLDzaVVoLDTHLzi9ejpAkUIXgEFaPNGzQ8DYiL8N2klRozLSlnDEMr
xTHuOMkklNO7SiTluAUBvXrjxfGqe/gwJOHxXQGHC8W6vyhR2BdVx9PKFVebWjlr
gzRMpGgWnjsaz0ldu3uO7ozRxZg8FgdToIzAIaTytpHKI8HvONvPJlYywOMC1gRi
KwX6p26xaVtCV8PbDpF3RHuEJV1NU6PDIhaIHhdL374BiX/KmcJ6yv7tbkczpK+V
-----END RSA PRIVATE KEY-----`,
)

var testKeyFingerPrint = "b4:8a:7d:54:e6:81:04:b2:99:8e:b3:ed:10:e2:12:2b"
var testTenancyOCID = "ocid1.tenancy.oc1..aaaaaaaaq3hulfjvrouw3e6qx2ncxtp256aq7etiabqqtzunnhxjslzkfyxq"
var testUserOCID = "ocid1.user.oc1..aaaaaaaaflxvsdpjs5ztahmsf7vjxy5kdqnuzyqpvwnncbkfhavexwd4w5ra"
var password = "password"

func createClientForTest() (c *Client) {
	c, _ = NewClient(
		testUserOCID,
		testTenancyOCID,
		testKeyFingerPrint,
		PrivateKeyBytes([]byte(testPrivateKey)),
		PrivateKeyPassword("password"),
	)
	return
}

func TestClientCreation(t *testing.T) {
	c, err := NewClient(
		testUserOCID,
		testTenancyOCID,
		testKeyFingerPrint,
		PrivateKeyBytes([]byte(testPrivateKey)),
		PrivateKeyPassword("password"),
	)
	assert.Nil(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, c.authInfo.keyFingerPrint, testKeyFingerPrint)
	assert.Equal(t, c.authInfo.tenancyOCID, testTenancyOCID)
	assert.Equal(t, c.authInfo.userOCID, testUserOCID)
}

func getTestDataPEMPath() (testDataPath string) {
	testDataPath = os.Getenv("BAREMETAL_SDK_PEM_DATA_PATH")

	if testDataPath == "" {
		home := os.Getenv("HOME")
		defaultPath := path.Join(home, ".oraclebmc", "bmcs_api_key.pem")
		if _, err := os.Stat(defaultPath); !os.IsNotExist(err) {
			return defaultPath
		}
		testDataPath = path.Join(
			"testdata",
			"private.pem",
		)
	}
	return
}

func TestPrivateKeyCreationFromPemFile(t *testing.T) {

	key, e := PrivateKeyFromFile(getTestDataPEMPath(), &password)
	assert.Nil(t, e)
	assert.NotNil(t, key)

}

var testNoPasswordPK = []byte(
	`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEApX5BXxc/tmHFCpnad2NK3/pVccxf6CK12NCPT2YZjEgQR+9t
GF53RFHZlXOXb+aahbPE2O8gupbVur0Y2C6rqzRZpFUcvM+QdbGLL4WPQ+gtYrEC
IJRarQ5UWhgnelPI2kzw3RvZS977Fdr7QX19xIbeQqRtOH4eOkCJZgJuufM0Q6Yi
RWSHjzNYsvVbETfLsHOLNRN9xOYcUWPMi9qkGpytlV1eyiCAMYlxfjsvMat4gMeh
GPz6Sj1OJ8ZuT/6Hdy5HPKoyPQHwZpWq0fiOniyhnyp/1BjKKTHPdFL5Bt3dzk7h
nQPdfe6gSjbMyzbCaOrCvKY94V6VO+UzVf3SWQIDAQABAoIBACapw34CwXjLiKw8
W4TO5rxDENlARRvHmDJqL0D+enOCloMn1ZX+4+BLOwkmczfKaUlZQWDpJP1SpeY1
rWs8JBEgbtzsoYUe/QHyE7Frg5f60zeeYP/ZiQGrOlu+DuMOVftiRFdz3SVTl9d4
TID1X3+dfqmVHos3M7qqPy9c3B+G6P7vr+HEh8hQpkWtEqZQxuqp1mdWhG732x+A
//KGUBzdS+GELEgHBxf71jS2VhljPUJo9zAWSoNYeXbQUmHN7dS4Ze6rL2Icj4WL
KQcyWpqXqsT9SwjWWHfcpm4CqXaN68jcm9qouyaurkdVo+hF/g8K1GQvM44D+HOP
49LwfI0CgYEAznrE8MtarPPKacCdczOxr+1hrXj+mgVvemiUrgHACH1Rt7cLzRAB
B3LNH/uo6RzcSVgLRqTf7lb9eMoc52g1njgoAMHa1CgRf3QMiP/+ssBlf87SGx2X
u4GX4NiN+KJ8rRpgM8mxiJw0+uXA0jNJzaKFKip5fa4RIjPdo7aDR0cCgYEAzS8L
UGJn/r3Yi9vD/r+fEUm0ENhF2YHUIRNkkmZ8aMQTjfk0GGL5HStd3Sm+4skpj85+
fvxj5WIaatI4kDdiMINBSTh/HFEqZhxw+pbQcl0uBVy1bFF+AlQdx/lmaJZqjDtX
UUrgZ5Win46QLNJhY4LRA8eI/9mOt2hDoSpEKV8CgYBQhCdQDrxpPRftbSL4zWu4
wsSYNNpzjTMPdMClqiEMLnIzRbngWSFNmkLK+gPAA3UTVLXw8lIwStPEymvDASwH
araOtQl0Obu5C7PnqIvVgJkT4b6kvEFy6PIkx81060fa6LIi/7+vGdq/C+DJFx7s
hTeQXcfKbppX0AnZ0U4X+QKBgQC6Qg5fNjV5RhUhQKo2wvQ+2U0gTXN68yQBsn0F
eQtOf0/Q/XuQ96d0Fz3p2k9xx3J3HNgvpiV4wQmCFrtKDzyPFVdahHK+3d9DOmZE
1Er8xiFUtMfsQD3HF1zBf2C7aG/oRKYLIZF79pXdiajPR1so3kOmzqdKuc+YJond
72RYuQKBgQCOSibFsiwqVdcFEAJcyaFeAt9OZPtlZXFd0Zk5G/gzTgGfHeDj3VKw
Bt0h0Tcgv6ERGXmcOijX7Bg9IpKDhb2ge/G/s8RQAVuXdWEcXbpg6voHVg+vKnZY
sAlNkj+qwrHRD/Xz6gREXXjbV9I0CKY4oxb9Ineu3BAxGS0J8uY7Og==
-----END RSA PRIVATE KEY-----`,
)

func TestPrivateKeyCreationNoPwd(t *testing.T) {
	key, err := PrivateKeyFromBytes(testNoPasswordPK, nil)
	assert.Nil(t, err)
	assert.NotNil(t, key)
}

func getTestDataPrivateKeyPath() (testDataPath string) {
	testDataPath = os.Getenv("BAREMETAL_PRIVATE_KEY_PATH")

	if testDataPath == "" {
		testDataPath = path.Join(
			"testdata",
			"mock-ssh-private-key",
		)
	}
	return
}

func TestPrivateKeyFromUnencryptedFile(t *testing.T) {
	key, err := PrivateKeyFromFile(getTestDataPrivateKeyPath(), nil)
	assert.Nil(t, err)
	assert.NotNil(t, key)
}
