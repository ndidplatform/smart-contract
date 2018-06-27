/**
 * Copyright (c) 2018, 2019 National Digital ID COMPANY LIMITED
 *
 * This file is part of NDID software.
 *
 * NDID is the free software: you can redistribute it and/or modify it under
 * the terms of the Affero GNU General Public License as published by the
 * Free Software Foundation, either version 3 of the License, or any later
 * version.
 *
 * NDID is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 * See the Affero GNU General Public License for more details.
 *
 * You should have received a copy of the Affero GNU General Public License
 * along with the NDID source code. If not, see https://www.gnu.org/licenses/agpl.txt.
 *
 * Please contact info@ndid.co.th for any further questions
 *
 */

package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ndidplatform/smart-contract/abci/did"
	"github.com/tendermint/tmlibs/common"
)

type ResponseTx struct {
	Result struct {
		Height  int `json:"height"`
		CheckTx struct {
			Code int      `json:"code"`
			Log  string   `json:"log"`
			Fee  struct{} `json:"fee"`
		} `json:"check_tx"`
		DeliverTx struct {
			Log  string   `json:"log"`
			Fee  struct{} `json:"fee"`
			Tags []common.KVPair
		} `json:"deliver_tx"`
		Hash string `json:"hash"`
	} `json:"result"`
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
}

type ResponseQuery struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Result  struct {
		Response struct {
			Log    string `json:"log"`
			Value  string `json:"value"`
			Height string `json:"height"`
		} `json:"response"`
	} `json:"result"`
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = defaultValue
	}
	return value
}

func getPrivateKeyFromString(privK string) *rsa.PrivateKey {
	privK = strings.Replace(privK, "\t", "", -1)
	block, _ := pem.Decode([]byte(privK))
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println(err.Error())
	}
	return privateKey
}

func generatePublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	privBlock := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   pubKeyBytes,
	}
	publicPEM := pem.EncodeToMemory(&privBlock)
	return publicPEM, nil
}

var tendermintAddr = getEnv("TENDERMINT_ADDRESS", "http://localhost:45000")

func callTendermint(fnName []byte, param []byte, nonce []byte, signature []byte, nodeID []byte) (interface{}, error) {
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	var path []byte
	path = append(path, fnName...)
	path = append(path, []byte("|")...)
	path = append(path, param...)
	path = append(path, []byte("|")...)
	path = append(path, nonce...)
	path = append(path, []byte("|")...)
	path = append(path, []byte(signatureBase64)...)
	path = append(path, []byte("|")...)
	path = append(path, nodeID...)

	// fmt.Println(string(path))
	pathBase64 := base64.StdEncoding.EncodeToString(path)
	url := tendermintAddr + "/broadcast_tx_commit?tx=" + `"` + pathBase64 + `"`

	// fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body ResponseTx
	json.NewDecoder(resp.Body).Decode(&body)
	return body, nil
}

func queryTendermint(fnName []byte, param []byte) (interface{}, error) {
	var path []byte
	path = append(path, fnName...)
	path = append(path, []byte("|")...)
	path = append(path, param...)

	// fmt.Println(string(path))
	pathBase64 := base64.StdEncoding.EncodeToString(path)
	url := tendermintAddr + "/abci_query?data=" + `"` + pathBase64 + `"`

	// fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var body ResponseQuery
	json.NewDecoder(resp.Body).Decode(&body)
	return body, nil
}

var ndidPrivK = `-----BEGIN RSA PRIVATE KEY-----
	MIIEpAIBAAKCAQEA30i6deo6vqxPdoxA9pUpuBag/cVwEVWO8dds5QDfu/z957zx
	XUCYRxaiRWGAbOta4K5/7cxlsqI8fCvoSyAa/B7GTSc3vivK/GWUFP+sQ/Mj6C/f
	gw5pxK/+olBzfzLMDEOwFRbnYtPtbWozfvceq77fEReTUdBGRLak7twxLrRPNzIu
	/Gqvn5AR8urXyF4r143CgReGkXTTmOvHpHu98kCQSINFuwBB98RLFuWdVwkrHyza
	GnymQu+0OR1Z+1MDIQ9WlViD1iaJhYKA6a0G0O4Nns6ISPYSh7W7fI31gWTgHUZN
	5iTkLb9t27DpW9G+DXryq+Pnl5c+z7es/7T34QIDAQABAoIBAD/nq941tKx/2ppe
	V/V7CZ6zc05OZN3BNBFJi9QbJO3D4dOigx4ib7Lg6n6bAkuqLK9joh+oQW8X+eG8
	G1btEGwaTr0kPVMDa6xDUleUOXSVMTCyCvGSfXkaufEwv22nVzknYk0W6hCiATEw
	lR6Akdmr3mIg8jwXNRVThO8MPFNWJK2TEKM+VYyRJaHrTiVnnGBaAc+6jM19xh9V
	92j0O/+wN+XvOt0m41+PZxz37nKRqX0HVqo/RZJ7OwzyGtPdNMd9AlXftS8eQhhG
	GopFPuEjWDjAIziC1MBtI23BFAp7cb7hkDK4p0D2ZZRrcGizA3ah/hvv0cUGaBb1
	EMzJmvECgYEA//D9aUm8T62OZzNdCC0lxPx/tS3kG8tf1hMGjU1zT5JResCZCvsk
	Xd1PS/62EWg5KHgz4Vn1eOApbYtPDiKOSiZAj2/pLhvRpStukC5I7ITDnE0sgDbt
	I/kzfcGR8TsVZVHu9FIvoZ8WrzBTNwC1uihOpxFpVOS23365fUXP9t0CgYEA31XS
	v2pjBpP5a7TFzwz7ULSBzO+sR1Qm++W7bUM79DJfhiB1piRordhxjERL1W95FX2z
	3/V4bQ7ophPkfcy68RCC5Wfts/+lHQoVNlG6suEsmBsac0g7ONcsaMrEvn9JBVxE
	g6bO7CRRHZ9o7nl4k0nWEbxhyraWHFbqBISQutUCgYA8czf3QUIn848Z0ujbQIaW
	Mykaqt8grXVSQ6Ydg7iDh8SU4J6FGHIrdVUAVwW7sMknRNTEGhI/XXqLdAbVCNZg
	rw46kq0ZhdqLT2nKxhPVQTpOVW/4TIDQKVC/GBQXTOQtzR9KN4smekPKVvigmhtR
	/6ksDpG5SlfjC7RV4UJQRQKBgQCubwba0Iolkh/GWvwAup/zqfiTi0Lgtz53kjgw
	n8nM8icfyGx7ZoaH+by+FH2yZ42IFpUOQFhdvb5CMNlO1D/SltXVvbWv1+UraDun
	IHCU1ECTUN/42JrAy3b5Jh5Ct4Hd+PHebcPCNp9QZrh7Qk7Fo27ajWtH/BIEcnH3
	M18jPQKBgQC5M1E9aUYdVRC2j6HyamAm73uLOdltQ7S0pKtypjMYBEUVzbXet2TS
	iEkC3ntWGFU9RAOKvwFWvOz3Vuxqgl+H5nkoYH6qhkgAuqOGAHb/DD4VHWzxwHnD
	W3SgZHxUij7PPJ8Vslvoov9SZIq5vBZiWvfNKOb4/8KD1IK9dO1aKw==
	-----END RSA PRIVATE KEY-----`

var rpPrivK = `-----BEGIN RSA PRIVATE KEY-----
	MIIEpAIBAAKCAQEAwCB4UBzQcnd6GAzPgbt9j2idW23qKZrsvldPNifmOPLfLlMu
	sv4EcyJf4L42/aQbTn1rVSu1blGkuCK+oRlKWmZEWh3xv9qrwCwov9Jme/KOE98z
	OMB10/xwnYotPadV0de80wGvKT7OlBlGulQRRhhgENNCPSxdUlozrPhrzGstXDr9
	zTYQoR3UD/7Ntmew3mnXvKj/8+U48hw913Xn6btBP3Uqg2OurXDGdrWciWgIMDEG
	yk65NOc8FOGa4AjYXzyi9TqOIfmysWhzKzU+fLysZQo10DfznnQN3w9+pI+20j2z
	B6ggpL75RjZKYgHU49pbvjF/eOSTOg9o5HwX0wIDAQABAoIBAQC4cjOvDYqcadFg
	J2RLcvj+5Xs0HFiSqrYfoehc4H8oKxpR+e+6TR1ufxC2zUYzyQmiF8wkTzr19xGA
	6XJDbOkx0j5KmbbN7hu2+W4Bgfd7hQgbUct172bvJcnjpJT8PJqqQ0h29oX3veFK
	0t1Q4oZW2e3YGUjdO6s39Xro0vGCo7GbUlQSwS93sDQoGhrVmt4hsfjCpXY1+cLA
	EY6ZMx1d+R/cMP26AxHrKpyum0VD4cNlsBwbk7gJY5rLCNpMPLNj8ns/0ZjuWRir
	1x0diYH1FpoOoOiKq2hU9OdYE6DyNGeTpxSQODKEUwwSEpCJJwKfONp0bnjDqaLK
	d+0+LEupAoGBAMWyld7Ej8Au9+w6VgiuHzJHK87PUF73GjAbDmVQJidf49YSBjSS
	fOzVAL2Wy4l/fKC73P7mb1U2DyYt3v3ywqPoLsGVg82aN3/btjOpEB/CJKEQct/k
	IYOQ82+81MQw06jBv5eB7xjpH9wa6t5lafV32XWLyHgwYUzJe4gm0ltdAoGBAPjJ
	UCqB8wHT+hGDwqTM0np+x6uLYSltMPK9HjMr35VvzYpIVlk6qZHxIBHzjg3AsC9V
	v65owQB9wt2hsdY3R04XlUuIeOHSZs3OLpacmiqtOsLa53RoZCxSNlCu5NGIAyer
	Yxfwx4IdHO73gPDnsYLxrf5bSlL3qb6LMXMPezzvAoGAIl1or9B7LGz9q5J4Ygni
	Ylr8wnZHAjrx0mrhlbrY5v9EG3IGohzUmlZsSohr2PrQLyB4ydZEhAthlsFigcIx
	E0zI092pi5PDEfafNVut8ddNhrHVRhXhvXz00/d/BJt4L11+cFeluC7N2vTS3tXC
	FWk/467oqfu+7hoX3xLgfgECgYB25AW6eqV9zyZnLldbWEKRpXqYISCKopLMvdHr
	1GCh0m8gUVdqht04UEnqKkFNkzLfPBRBLfBl4rO4JKiO3ZXm3OBM22ghSuI0If8j
	nK0UDfrR2bjYaXbNs3AfeKUC+QPA9meBrmA5bt4/2Om2tpKfKA3lSw0mvxJQa8Zy
	3Qgg4wKBgQCkFFf4w0aUUVutEI70B5hebrq37ktj9afpntVoLpb220u7NWsMdwps
	2L0EW0Y2pyszOv1TEwgAVsKGcpPzN3FJk861fv64cpn8CIzvBwoA4UWAoOCGkJkD
	mtlAVXOCyaIG241q6LvHEcY9S/oc2DtPpnCJB4HiAf3lDLd81Hzkaw==
	-----END RSA PRIVATE KEY-----`

var rpPrivK2 = `-----BEGIN RSA PRIVATE KEY-----
	MIIEpAIBAAKCAQEA1QXXrV7X1b8uFL1PW7+FimlAwxwbEMG5hFru1CN8WsRt8ZVQ
	IkXRpiwNNXh1GS0Qmshnv8pKaNCZ5q5wFdUelYspZHVRbIkHiQAaEU5yG9SyavHs
	DntUOd50PQ3nC71feW+ff8tvQcJ7+gqf8nZ6UAWpG4bvakPtrJ81h4/Qc23vhtbc
	ouP0adgdw6UA0kcdGhTESYMBU0dx/NNysvJhNx36z2UU6kbQ3a2/bINEZAgLfJ7/
	Y+/647+tc7bUYdqj3dNkbnk1xiXh5dTLsiow5Xvukpy2uA44M/r2Q5VRfbH2ZrBZ
	lgf/XEOZs7zppySgaTWRB5eDTm+YxxyOyykn8wIDAQABAoIBAQCd1/ttInbJkiSi
	B3hzImHgIodzSzMe4n0Ffp+zHyw40Y4p0RqUmqly+Pc8pKoX4pWIK3D84vbp3Y/8
	J0s0UjucUYZ1Qpz30D1+HU4zfq38w0kFB4eDX40UaCo3R0LpJwREphpIhkRFNMfK
	ie7kqTeObfNVS1HBqt3E6B+w+DZcIEI9phmrOcnjEAzPDI4q4sDIUhpHv84tkb/6
	lm1RWDlRxgDOGv3knUVXaOvAkTRqdBINKOhaS6dLPpN9FL9aj5UKEklxEtoPSaFP
	ib2+RWWe4B+0FPEg0zuSTIH6hhUQK5CBa3CM+0WzfZmsqSpYFbCrmWeOT33tPGy9
	NlgVQfwBAoGBAPUdTCFyJPrYdff6VDfxvCDMeLYKJckaa+l3M2Du5BtgFQJM+yPw
	5JkNGUyF9MFNWX47cBm0W7pEU8IEiuokhF9XSizX/H8Tz9YyixIVU4krCTtyR6bX
	xl36KsB9t8vNtXSN4M8VMMlAZWp5q/n36rzy7jpQKkFfq2todd1yGB3zAoGBAN57
	sigfnvxIm41SjAxXnvY8KoP0jCTBxmlxgvLsFhpj9lUIZqQmxbYgeggI9c0MIBNa
	/QzmrzLHnSSjtqoUxXy9XY3WKE60uHcvPzePKw5V7EdBWdZLWO+dnKFMziLg1gkx
	ccXp0T7VtdQenKRga0PGWw82X/Sr+h90TTmPi04BAoGBAK1Xkb5ZZbOMHylGfAaw
	SrX7RCag2IX2zHfn14rmhqShd1oQLM8HDfL643hNh4CoffCagjV7ah85MO6VndPm
	DUMLjSZXfHY2AZZeWiFouZHYwIes0uU31U4im9dTUQatLHUH3QM13jGE+/Onpip5
	3CTRvA27IZbn3GdyEWCQzmNnAoGAY+laxWgF5rfYmyuB1x0WNvAoC6Aru2oF515h
	dyQMfQd9HQyrw3Xh/fsxsiAL+mxCj06iK0QBU6WO7WBT7KdtVKpZtBODgGzqFiPy
	mMnDhSmS9SDk7jZiFyFJsKokPEeJ9xDsTfvFyxkAEeU5ZRwjr4kJZZh+mQsORUfe
	UkYjQgECgYBEXdVe0uv3V1dUVATfGdBCaZ6BxJVO/VGDKNfoEtFCkR7rVsb+kY6u
	DBW8CMBlUhoaSE+/BBEAyyzV++j0nC0cnlU1694HcM0hKx35F4CHPKEAH/ChrTVn
	EClJgxaTPuL9ON4s2OaQevT+STkx/dWH/O1FkY5oTAR5JO/wbPZqoQ==
	-----END RSA PRIVATE KEY-----`

var idpPrivK = `-----BEGIN RSA PRIVATE KEY-----
	MIIEogIBAAKCAQEAwx9oT44DmDRiQJ1K0b9QolEsrQ51hBUDq3oCKTffBikYenSU
	QNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9cNBMzSLMolltw0EerF0Ckz0Svvie
	1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZXoCpxUPQq7SMLoYEK1c+e3l3H0bfh
	6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksmz1WIT/C1XcHHVwCIJGSdZw5F6Y2g
	BjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69b0FgoE6qivDTqYfr80Y345Qe/qPG
	Dvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7jwIDAQABAoIBABVYHS/+p/QBXvIU
	gre5BtgKqylvGHnPVqxuV0gs/W+OFUhn8kO5r1ArukwBWXKqKxZXpH1Tt2VXoKMi
	NBznwzmQ/6W89ceYosImIHXYYsy6dI+BYNbdWaz49g7VxikXFA03WmZWACYIRwwW
	UQiayiESI30oiH2SRNZw6D+FS6qlRIR40M6GL3PS+iB6jPslXGencwxhS9jLQopF
	47igwzGyO4UZ7eeMG3Nn4IfyCxhHhXqpCE2J0XQIMBtXb9O//tvuAcL3db0jGMTt
	v8Q22JxF1X+0/1VfU9JGKfqoqRcOXqix4ztgLXKFel+bj7m64T08MKljD21k3bb5
	XrEDfpkCgYEAzg6406B3Peo2kZT5uVh3WzsFuVKAfko0H/vmMXMWM569LtEndjQL
	hib4hgkn9co/FBtliC2w03wq/9sKoIPZn73H5EkaAOjyOnF/Z2xvwj/irMJLov9g
	i5eHQJljWktzs0aqv2N1r9n6GR3uJDUya07rTpIkWNh3nn5fuO+srjsCgYEA8mo2
	CcXGvBeFNIcLR9SFaqNaDPirgoqNmrtxRoaoQvvWVU4rcnXO28E/0vn4l2OUvz5N
	IAC+3zEcSTRI//ZGKU6tzJXTbj9BRflLhwhr78D6ArGtyjgWXPBUSp71+qLSsRl5
	sHVSsc3acPpMRgGMrlA83OaMnArKGtCcL/qh7r0CgYBPw4EmYpJmBDj1Z963MZia
	VyGjGF2nBWBiFSeJcsxgVQ1UhyAocIMZfhJsCDVQvuZmCSjnaxBs/T7D5e2aLw/Z
	9yPeqbGIMqQ5nV+9EEu+vO4pA9k1knez8YcoqXe9J0H1XuCPz5dp6A4ZFO3vVCxd
	P6J0uruZLMo5LyAsvZJxqwKBgENxKzGS1ZSU0plnjMriJHAjnDUJpeW+mGDZD024
	vu1L1TiMc+f3QKLA4/nVU8UCjmqacaiarH+50Q3Ivxp/MMvjONU3RchhTs6h6dJa
	lHTyclv3hMtCyW336uuLyBF/5TAiT0m5ilUvWTufV0MOwU3pwtUOS0ZKdin5qcpr
	Z0vdAoGABAbnRGHFGBm4jpGhvKT8iXCoRlMvvaYalhpVAYAMehwLR7Chq3O6uTJm
	1/iYmfvDJP0ihXWbHJePpTQRnjAu0wOwxeFCS9X1dvFtJEt5DTRbyoZ2hZfwFPxD
	GN1apL6Q4DpjL2ktGkYgaKp6HW+5ogfOPPOJOICDScu5Ozl29+o=
	-----END RSA PRIVATE KEY-----`

var idpPrivK2 = `-----BEGIN RSA PRIVATE KEY-----
	MIIEowIBAAKCAQEArdcKj/gAetVyg6Nn2lDim/UJYQsQCav60EVbECm5EVT8Wgnp
	zO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2SA8Xlf+ZDaCELba/85Nb+IppLBdP
	ywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7jiU1lf37lwTZaO0COAuu8Vt9GcwYP
	h7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DGr/ZKzEE9/5bJJJRS635OA2T4gIY9
	XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15ycS4C35tjM8iT5djsRcR+MJeXyvu
	rkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpDPwIDAQABAoIBAFAfcxDUH3R9+J/P
	qsgmy6tSDxaZQLUUfS+NJ+GVOWVRpFXjh80bARm8r9Sy/mGQDS6Z9jJp8lDXgPyG
	Rs0OFl40BozuF57+Qq3HM0WSv6sYME1QGUjdIuPRyh8KfoxBw4MBUzvQ42JyYf16
	Xs/QKrNX5tYSqNaatI/muw2BlXb8Az/rTdI2i42BLHLDFHo9MS2+bAs+qg/O78pH
	/x+H0jNj/5ClL0k7i04WKSCoQNIcFbSOtEtk5ZlaALUWEZavgNn0t8L1ZAd2NK5b
	P5VYNwYiyrQhoX6fQkpUcCUwaVq6HnrWtDToBbZnr0pI2dtK+po5CSCLqQ1zr1pd
	zRyAlXECgYEA1r5k+n7kGgj5SM1p6vK2KxtnYQ7YYTwTslYxtO+Fy2GhYEqGYTyf
	Owe59grnv+RbINERQ3dxsKnC/3DjlS0+uarwfFdVPZYg/7W6FVGfn46FF30/SxqT
	AeejFGRgpYGweFHmtCCDRFDCT8BvH0mLRdGI/vMn5p0okZFQaVBG40UCgYEAzzzn
	nU1BpkeHMZuzQ+1amReN3Op2afrnr50ho3io8YyfTX2sJmyH0HnlxulJ8cB7bYmN
	eqIvXTwLDKPldXaY4NbMXQ6zanbFX/psIPxlRTf2NVqysBBKMAVQg+ZHAUMN0gIs
	xw7SCVTaGzSIxmq5FWvHidInf4Dphr2y4SuckrMCgYAE9E+QF+1bTGmz7ElNSlw5
	kmBINPd5BtHNg3+SFRSZJJ98gTuocqWZzwvTSV0faD1R/IDRdagB02jUS950Sp7v
	2anCtKEa0qPgQmkQpNlx7O/VIuaa7PoHSTjR957jMqLHo9wWu8lLgjF5dY8awa+c
	5MCsYR/Cik2tThT02Q1JoQKBgES99CpGlS897Md02Ur/8Zx0prcQAwV2l+G14pGi
	FZBCUBlZRYBdYdOyi5imi8OoUIjuJsL2B3YK07N2rkd/doimV5XKqZL4INKMc8+h
	SUpjnMTn9/vU+3bgXGvUN9tgTbZKyGWjMeKshciebXw7rHdBkCfUUQvHTC9Iv4xX
	dhFnAoGBAL+JIdwLcumdEZJyA3pBu6WfvQJzjDv0HE5N9FsLyTyqc3yacaxH/hv8
	nplQvZKsbhmtxbu/MGbJfp1cH0LgO5OamHj5TBEdXWxtZKgE2nmxJz0Fm9L8vZ8b
	H5plzde3fZP4YVOa+bK5XuHS5CrwjHoDItfvdPNF8D2rutrl77D9
	-----END RSA PRIVATE KEY-----`

var idpPrivK3 = `-----BEGIN RSA PRIVATE KEY-----
	MIIEpAIBAAKCAQEAz7N55vytQuBV17KHPzd1ILPonOpltFqcMCV+x81NJNcvf2Qc
	DDemYK2oObcs8rDuavx3+aSAeBrGXmFIjvVT7YTpEfoCGVf50AJKeyOeuaGefVy1
	2GlGUsxKxCWDJaWe6Vc7S+cOyiLHNp/U/La3rSRbJeS6+GLbbVtJZpXsJwIejrK5
	2JwSnCTH9aeVUDovJZNfQvPHaKArqermyI7/44o8qfGkImAs4UhLLpcQVyyADaqH
	MFKpRTE/cLISCB6Ut9Vb1lyBgk0xlGWLfrXa0erk96NK3tw0thd464qz2qFojNIS
	mM1ddG+VSHoZUu7UJzeCUXyw0RkB1PZEXiwz7wIDAQABAoIBACO1JFj8ycC8lqV9
	kNjibOWRaIVJmvCVv1Jbr98jwYZ65DSPfm7vRlBKqqg5gKW8m1CTVQD7Mgbz+3SQ
	XwwMy0ADYJpxk9jNkiobqrhe2FProDbHMJAjES785kGwfUqEnbxZ/dy/vYAs2Hjg
	o5pKw2sl2/G40BgRzs2PKyBS2AWgfSKoh+607mepFNKg0/Hhlvxs3eJv0h9//ez+
	himAAg9OY37gkWHS5DlQ7DsQgfFhRUjCFPGcB9Wu3fuMAnOwthtpigDfrp5SqvZI
	KrCwJKJkCiqpa7yd9qx4481zTQm/ZyASLiu8CXC9YVgLqhFcxwaOhh5Q0jdcID+Q
	5BnGzgECgYEA+MSae2o0a/P1RyBI6IzRo9zDhtTBFA5XZD3tstOoeh0Huhsrp/KR
	0L3nxMm9/EouibyW/SwyOVY3doxo9GNN4PWJg3p58bmYkG+ogwZD2VN1s8tAbU6A
	YCFtlz9xIx/12Vx5PKL8fr1FQdWeLleI+F3VtrZ8wS083wcbqcySWwECgYEA1b09
	+Gv6EGNaGDV3PwbDwAUg9AHTE7TB8QwkB6wkrNN5MuQc40xQpG6xzx/2kf89qEqX
	q3gBaaVvFEtH9QCa8sjqix6/1Rs4+V8D8lfjva6FRBtgm2Yaovhrb/ew5npb7KFm
	nz/cEUxJ7eXZ2QJJMsGCC0v0OPrlqkCXycp/Pu8CgYBnrfj8is0CWRDW7fu1AEu3
	UaEkJrO52ihOHQleSJylGEhKJlzRiGWBbESWXcaSyZAP08vSBIOCJg7Dl81+XYzt
	vyfq5jbAqiuNtxuyUAAjKYeawZE+fUM/zW7RZJ2QmBds2f+laAB4CgY9Y/yjL9Rk
	Pyd9GR1xnZsLEPlUkXBGAQKBgQCpzI1OrXkbS9JnKRJyn40jHu/u6QQmw5LPTDXT
	Yo5APkAqjc3lRNtLxiS7x0i682qoJ5oWPl/g7eww0x13JePyvGqX2vXK9rVsZm9c
	NzZVmi+Ey7sTuSmwDmpLqRp//vTIJ/C+0pyhoVmaBN/r5kUAbXpCPzTlj2yktGvh
	g11TQQKBgQCtUc8dgxRBEAnlCMjjhVK+8vyl4Tk2dcLL4U7stk+3hstN6bJUFWhl
	lsPD7SyWCfa6BdAs9DsTLdUa4EGvfVkRn6oar8OMC+OMhDethRjUmIJV+wWS3ati
	I4EPHrPYK3GNb75+G+qH9uJZ1e2FM7CGaDiBHVSthBkCqjEv2e5JGQ==
	-----END RSA PRIVATE KEY-----`

var asPrivK = `-----BEGIN RSA PRIVATE KEY-----
	MIIEogIBAAKCAQEApT8lXT9CDRZZkvhZLBD66o7igZf6sj/o0XooaTuy2HuCt6yE
	O8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQrdqGwpogvkZ3uUahwE9ZgOj6h4fq
	9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3ZuiD02QknUNiPFvf+BWIoC8oe6Abyc
	tnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU3pqukT35tgOcvcSAMVJJ06B3uyk1
	9MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gVFt93/0FPOH3m4o+9+1OStP51Un4o
	H3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kjdQIDAQABAoIBACUI4fbkFomYWLr3
	rgSSSaoIG/uvdCA+8o8AMc5j8tFR3RoNMBW2Ep1Ah1QYfpPnS2zndO0FqnKmjvWM
	nY0EUyijsVAr+uqqIGsFyXqwTf72OC/n5mEQxVFQ9IyOb5npPuMRXAU8upJ+5HAZ
	HGGvyVX/Ygm5QjZgDhFnEjYluENii7wbYC7YfHsdkNoUH/10Y+xnfiPpd1X2fD4/
	yV7NgrO9uwsLJMHjKELusZwMJ7lQ5JeSHxuTfCKJdyyuaqSij9Q46q/jLuU27Te/
	0LuHRl4zggtDRGXfVL4kEjU3B6uXFg2CXFmG+S7mb4zjIhJvTbkj0X7ZrR49kcMB
	4btdGAECgYEAze5J0GAPTseiASSN3HI6OIDCvG6iDVpjVszKLsHkfwLDRtr7kuhj
	FqAlQjC/dGzvVEWkBC0qZObk2R03E+tI7WdnWsiK01fwbZm7DTtFu9r4kGNvviGK
	FNQIuoH8YIn+XMFdjkuzzvw8SorRwWiKiSqDIAyCETdGVlIiuqOhwhECgYEAzWyP
	3XL4q+jY8CRCrNczO2lRoTDOH2zpkRxdoF795pGnvXM5M9KCOrcoFjdFPijpgZGR
	oIe6hnuipvpIGN13ycIXwa/qIJlALiAq5NSIK/ZKvCB3f6P+gucNs0b7yCceL1qa
	A95Qa8eAIoDBAkv1nF+jTZ4DLQAgB6R+g+W/pyUCgYBtjmIyu4gpT0e+9+WI7DRR
	Lx9rBCiulfHXkefWbEzVzXB6V7ITfBKLTPPFfQ2+MN46pToXBrhRKg2B/Gr66+fG
	dYak46AHw/cjN/Atn+T/hgVLO7uNGWbOoedq4hCUg5WRX0YYl+m3KrYgqi3hiW56
	fuV3vW/NHO0Mq3HSfY9nIQKBgByG3++jwK621jF7B5tTAzVT6dcVnPo2OLVDGClm
	J6I2RfIEJ0RwDk+zEakMIdyA9/RbT7rYPmngj3TautphHvpwrrXiBQRj48rEAtDm
	Rsa8HCLF63JZRsXM6lUkHWDtNb7juRGidM6S1NN1x9fWzpPZoCbuM4izRL9q83rD
	k/rVAoGAULKNeMhshK4hghLwYRrvKK+RvqTHzGRRilWqOmVOD5qm9VOpUlLNudao
	IdZYlh07pA1L+IXtGdFHL4GlTNa0xXQBsLOTpklqIrTC62ou6026ADM1SC+K/5GE
	98StPl4dYJRYvWjfKjfSkqI1J9pV6EPRIHwP+r5gB/EsBqKpmhc=
	-----END RSA PRIVATE KEY-----`

var asPrivK2 = `-----BEGIN RSA PRIVATE KEY-----
	MIIEpAIBAAKCAQEAzhJ5PP3dfQtpw9p0Kphb30gg9jpgsv425D5pzZaH00zPgYfN
	TVZWfrLlTtc/ja8dbHvyDaCyzFD++Vr1vtmSSs9/j8ZhTJrTYHoiHvfG1ulTl1Qd
	gwOcrKhpfhhjnCVCPOYjptgac/KPjhT7uiuYwB6axafx+RqPQqwQQhmuuxmTyy69
	l/cqezDtYCYUJVA6nV29ZaaF1VjWoE05PK168mcB5quBdE6Vkc4n2k0wxaaTd/s9
	LPy6STXtz5IBXH2Gy5RP0TGeXO6iur/ZSM2z/3vQkTMjY/mkDduGioXcB6ieNgVv
	3XYbZg4VJEDSuOpRZReKcgLXvwk3CqZZdZRRjQIDAQABAoIBABfdn9jmdc5Tkg4y
	sJ12Q72aNucNX8GbG3RXnh1HP7fC/4060xYP17iYs2HsH9oi27+Co0fcwphTERSD
	6k4OGJk9asKV8RLUI4La4jS/8XFWWG4AOeLAelassnr+DBs7XW58IMjj4jxnbSTB
	XV30Sp6FbNtTVfzJjKnmD4P4QXo9iKOqS1XHGMJ4cb3gizbtvsj4e2U3zVE5N+Um
	zvzXcpvyTH1yJrEY2iiorYFUxDzccuWgyQPTOP0Rtn961JlVFVRQmzyf7briVPYJ
	s6fjQqt5CH1pTiX1PRzJJx3Bef/6QAsogYDcYH+zZ5xz4ZZeRgPEVXH8tfonUr2E
	whOdAykCgYEA/OQcQP67JxjXmRVo0SLSNA6G5fcVwDp875dfu6cnMP1LqNq6NW0w
	D5g3prHPYZBz5556zKVitlWsjeq3ABCsxoknwbRBNPEMCWPG+7xi8ie/va8sQWqT
	W5vFWlu16difrpCh77sJhOrGV7jwvANmjg2ltiPkKwte9bWIdipKyl8CgYEA0JsE
	ODHO2XC+ggRfMItUndccT2GBgi48IEslnDiwVU+JSc9YKh9IxtFa7y43VMkbvncX
	qClJ2U78IkPs4OlFbftk0AttqvZfZmWnys8rMZgzNFOhtjgnyBLlRPn5K6z4EjQ+
	lEIQzNl0JmMq6x+P5HYxgHHvL5tncUQia9CdA5MCgYBiEvUCH8fk+bVjIPJtaNus
	ZJXcSV6eFhCtuj7eP4zratAUw/7DCX1CDv5GH18VrzfD86ocA2es3rz0rLobxFu9
	AyPv8z/2kCTi31cj+YNF9jReE7lOBU7wkBCRYk/CSMhkoqKqnhaq/YG+M3Lo90im
	fpRtdq3eI6LIF4a8jNpEcQKBgQDGm5A84E8L/qeiqf7m/QCm9nLhsPfYtaRRKrq4
	LdDUqFERkPNjxz1G7XQiXGIZuw9LG5/OXuEMoIK1LO6OhAmyWLL20KqtJrxVhVtn
	YC7DnSDDJQzFrFlTx4m5TjXJO3lD+7HI/c14+1/2XFw0V2xsG4utusv7C35E/JW5
	CHk1OQKBgQCdbhsGrBII6OaAYPKSLSsBQxczbvqE0EP5pH3E5IXKgrQFvkMLBm8D
	t+QcA95WPK7J4AEAaU994SeUT1EzHCEU6orxYeFMdSHn3C6/HI9AtnftIPT4lPih
	S0ya4kkk1gex6wejZdIAfSEoxNWJd//t9ERfGdGQVOOPsLiiu/W+Ig==
	-----END RSA PRIVATE KEY-----`

var allMasterKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAukTxVg8qpwXebALGCrlyiv8PNNxLo0CEX3N33cR1TNfImItd
5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlCBayeMiMT8tDmOtv1RqIxyLjEU8M0
RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNraI8YDfbjb9fNtSICiDzn3UcQj13iL
z5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV+s6b3JdqU2zdHeuaj9XjX7aNV7mv
njYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBOez6GqF2hZzqR9nM1K4aOedBMHint
Vnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03XDQIDAQABAoIBAHOrRlaTun/XlCRs
oICeYnPQKahAuLOa59jCQogzbEgYo5eey+PDRBKlJa0XpQGcMMyQnF9w1TRlBg64
SS2fakFNzTsDL2sUsDOSzcBcDiBAJKKDUJyHsbE0pxdFDi9r7QaXcrtfRqkKFV1U
zB0NsnKOjzJuLiGQVae1QW3FKEDcRBqjEtfGkxpSjGW+zvz23RhMIMfEgdezvtJM
dl1j+u2scy39m1fZuc4bU3fOoppe5LxlqazkVu6WiwbE3dN6QfJhiR8nXMg2RHIh
A5PuI+iqLmmiFW2uu9v4/8ISu/lMlwWU6rTJ0zX9ixO4aonpp6H7KCSmCmXOCM+E
WtMmDGECgYEA9G0O2oYNP+ZN5hbv3w95clbfnIoRZPpBOD8H2jbGlXdowUqQL72m
MfoFpq6VKYbDxPYzPfjZUoeta13bMgu3+4h+LehIR7uTjPvLTFH60Qr4z85RH6AF
SeOIjYBfIaVo/W/7KH0ezVgPZNvq0SwZyv2TwVrkVtMRc6xK1fxC5rkCgYEAwxbo
JS0XwodK5BX4hBh2MxI3r8lx2VZd6DzngwTz2mB3BwjPnqS3HVMOdYr5I4UJqQPn
HTtRrEIIdynh81n5H5Qkw0DVzGVmpJSItsJBzr46grY9p0OZfYZyYmeV+N28JNEY
QPqpFZSI2oyCn7km/2eqj2YYM75zjoBBrhP7SPUCgYEA2kzcy0aWZs+mGy25JpuH
eBsms4SMbIcl4LpKpRXu3mc7ZAbYKAtVd6U5jti119TI3AyXT24FirQqqo20y0m0
FC6foximFYruCSiJNayyOil2dwJpabldf9R7jQVt8Xrt/gwZYNv+up8/gHD5k7+z
eZxobnRjIzh3ibwDSoJ2reECgYEAgKgWqI24YZ1/kjO7FMJdEQkumEstPbtrasDf
nNQjTRzY4la5NVJDQJ+JpZLlAru1xzS/sdNw5T0XAB8q16W6WU0FgY68cHNe4aLj
FkO9ym5Bf/pXZnt6OgH0ZVkS2nDApzcN26xy3bx7FEYdzt/4C+9919vokhdDdfK3
XennihECgYEAo7gsAeulbJbeLDv4KntoRwl49n1YkO5yKza/icSASo/Qpk9n8Tr7
0Tr4yud9RbBwpnMsuj0FBwS6xrED+sqRlhl4Hg6rdPcj8Jl9+WndcwLjDLdo0Xrc
dEtsboiI40w8gQBC7GvFQ9ihmYQDOyhOlPLCmc4w2Yg2nkyfctJ9kcA=
-----END RSA PRIVATE KEY-----`

var idpPrivK4 = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEApbxaA5aKnkpnV7+dMW5x7iEINouvjhQ8gl6+8A6ApiVbYIzJ
CCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaHOk1FfX6tk5i5uGWifRQHueXhXjR8
HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7XHdh9Mqz+XAkq4T6vtdaocduarNG6
ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3THkM9noP2AYEH2MBHCzuu+bifCwur
OBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jdAmbbqMyGpAsF+WjW4V2rcTMFZOoY
FYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP2wIDAQABAoIBAHnL21K7tQ7ymtOP
i1OiWLOpLsH3EYKWOImOWz9LSRvQZECl9a65wwA5g69pNoN7s/Z39cVH73X6VNYh
EIFrUqFz29eH2sOW/xUWC71jlPH2kBKM+5DkF0DPluGfdsH/PcotCA5FvA1c5hK6
eHr2cJwMVqWBIEQ+sHZrfPFi2NMiYl7RB0gxwt+CY6ezDUY7TOqdx/3UIDXQa9nA
PQ2GvV8cfKQfLl3rfsoF2ObO+fs6Kmsp3aYJxRVu8aaD+UKS+ljPsD+/IYGa/8Z1
ixXVSHscDrIfRR7LmYDXtOal1pwo0gdErNAMivPXwY1XZ03iHTm7pihhTFd86jDs
9nLRjUECgYEA2jC3sPHDLEeXn8tVMj8B5VBSdFi+H5tyzJbR5ef0QAl7vslbi5sq
MGXtFhRMDjeTwLBAa5YMnYhqHV53GB2zK95+FJeQy1rfSQsxZwIFL0L4wySP19EI
aXaUpPDgdRKS1SC05auiQp5PE2AxmIHfBc77HDE8iU5SiS2TOKFf+isCgYEAwnSp
0Pa/lkHqfEqe6I56Zoi28eUGwjVWc/USdNsM8sfCja49UCryHHHljm1PfgFX4IFy
Kzupm7l1cHQsFI3o1VYWxo/DvXKJ6NjGZLvhbWMLsAaMptcWrkMZ/NDDThuyg9hg
QE3cvNYUk2eoy4PKcyqPeXggYd7Ue54EO+scmRECgYBya5no8N+pGOIqqjbDYsdb
ugODgAY0DRDmuTDZoAo2isKaCn43d+dn+guayIoZ6otRQRyHTujOs/rx69gIjYqo
NsVnhxQnkEAHzhbaLfUKE9TggQvt4XDH3aeV17vdqR/XJI+44Yj15o8RWiCoGXMb
WK/W2PsmBizCQ2QxDm+GgQKBgEicd6zn9rKM+ppe4ufEDECtXGMHOnbao+W45aNt
CHC/1w5AufRtlOq6PRXqC3zp036p14/9P2A+6HONbchfFUpUUzziAh2D36trBuom
ng7SpVKdn3fNaVK5C8Mz0TohbY9+BLL+YCbDafuBAa69D6PhiKG7EZx6MK3YW4xk
RtGBAoGBAMMb1JaB7GF7Yia/8K98PCUeIYOKiPL2XVo0Vpf8z59cytvo/l3VAXXe
tMtaNcpcjJxooc3p9fVlZrU39Zs2fwL53qwlrsIhDFcGWwSCZB6Bcec+LD+MSTr+
rdDFuWTHd+msuASyCbJ2vX3SwnL7g/vapR24umMI6lrJqFGu2aSF
-----END RSA PRIVATE KEY-----`

var idpPrivK5 = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAu9+CK/vznpXtAUC0QhuJgYKCfMMBiIgVcp2A+e+SsKvv6ESQ
72R8K6nQAhH2MGtnj3ScLI0tMwCtgotWCEGiyUXKXLVTiqAqtwflCUVuxCDVuvOm
3GQCxvwzE34jEgbGZ33G3tV7uKTtifhoJzVYD+WkZVslBhaBgQCUewCX4zkCCTYC
5VEhkr7K8HGEr6n1eBOO5VORCkrHKYoZK7euNjyWvWYyVN07F8K0RhgIF9Xsa6Ti
u1Yf8zuyJ/awR6U4Nw+oTkvRpx64+caBNYgR4n8peg9ZJeTAwV49o1ymx34pPjHU
gSdpyhZX4i3z9ji+o7KbNkA/O0l+3doMuH1exwIDAQABAoIBAGyLLcILxy0QoeXf
ZEXtcvyIUquSXwhq1zlpFmNQrwezzt/6/WHSRItViQApMHu5EhQn4zM6PasB8T1D
E2mhwlNXJxt5B9NHxmYJAaLhoqVd8x4YN4eNoK0meLwCXHDFyUtxt7x2ywxa/YKB
Kmu8viwxGVIV3sYtqpTFqQOHzDlScRLv+qsgXw4OimjOTtvz1Gas7MnGnSweTYOo
6CAs3daORZ+tWwvwnAwjVIJItYC3T1hOOLJHkiB47EcKONjmt5jjy1Qvx7QRjK2c
/WOjGsO2tTAqfSW0jiEnybHTBqEqEKyufkDHKaChtur8sq5FFKdRaxnqKAfnQE3e
NaIMosECgYEA3lHLQT3kNUJk7UzxaBTlfmyaUG7pI4teFiPOCm6NQwQyiatPWSkt
Bo6uZ54r0Om2N6gPGl5rNC9M9njq0vPyOAKg2HT1E4LeoGPonEu65wiRdclLbtkW
NId7DYEtTCBI/imfncwSnMtLPTbnlZCCjQZljcuCk6A4hJ0SnxOS7FMCgYEA2FXH
zfcTt3rvTwMfrK44lxotOftlWbwu4nIjZQ57E+2JIV4eBg4hoOV3YoYVDb0nqBwk
5MGFJqrQoL9/1KbV/hkYZ4NDaCtT1U4iZ168rmbtQ1tO1QrA5D5IXjS0khC6syzO
716qZI1Y+xqQbnEEjEpzYnmwFL1PH6XFnqV+1T0CgYEArkV1y92lPy6diPrwnYML
5s9hI73dWXSNO1Oz1q+UYj0vFIXKPH0vg11jT2xIsooRwY0m0afD53NQpEBi6xw4
+jjtNuBvoGzM8POASsx+ZU5tH+S8EddwNZsiFZL2HB+OuFWOfpaS3H/rqb+ZR7+w
5rVl9AHciLZmt2WdTD9+w2sCgYEAqXwK3UIFIGofsjcwSYj0rOzFIffinzrfQGlL
cZC2vBYMqSejPfs0PWmI7pc9R1Y6C2qBPPaf6ntIl6dv7poGbNwcUnx0AthvBV4B
dhqyl6/rkimmySFznV1uNN/117lji5w/QylXNQ/H9nIJVX0VoxNw8mWDnbvykUi+
Wlwt0cECgYA9jlWS+rLppVflVTA4CAMNfXt8qyjDc3Wl3HUB0rgApRXbkMQb875U
6czVLqkWbRiIabfPXYrJRaTjG8thQuquzpaS6+U/ed6t0vUJuxTXkmj1HS3a/dJT
VTVj/BlsFEVvTc0wuiA3mwlgNirRI1UH0GYWlk22hUF3MpMF46SAVQ==
-----END RSA PRIVATE KEY-----`

var userNamespace = "cid"
var userID = "1234567890123"

func TestInitNDID(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidpublicKeyBytes, err := generatePublicKey(&ndidKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var initNDIDparam did.InitNDIDParam
	initNDIDparam.NodeID = "NDID"
	initNDIDparam.PublicKey = string(ndidpublicKeyBytes)

	initNDIDparamJSON, err := json.Marshal(initNDIDparam)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(initNDIDparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "InitNDID"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), initNDIDparamJSON, []byte(nonce), signature, []byte(initNDIDparam.NodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRegisterNodeRP(t *testing.T) {
	rpKey := getPrivateKeyFromString(rpPrivK)
	rpPublicKeyBytes, err := generatePublicKey(&rpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	rpKey2 := getPrivateKeyFromString(allMasterKey)
	rpPublicKeyBytes2, err := generatePublicKey(&rpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param did.RegisterNode
	param.NodeID = "RP1"
	param.PublicKey = string(rpPublicKeyBytes)
	param.MasterPublicKey = string(rpPublicKeyBytes2)
	param.Role = "RP"
	param.NodeName = "Node RP 1"

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	// for _, item := range resultObj.Result.DeliverTx.Tags {
	// 	fmt.Println(string(item.Key))
	// 	fmt.Println(string(item.Value))
	// }
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRegisterNodeIDP(t *testing.T) {
	idpKey := getPrivateKeyFromString(idpPrivK)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	idpKey2 := getPrivateKeyFromString(allMasterKey)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param did.RegisterNode
	param.NodeID = "IdP1"
	param.PublicKey = string(idpPublicKeyBytes)
	param.MasterPublicKey = string(idpPublicKeyBytes2)
	param.NodeName = "IdP Number 1 from ..."
	param.Role = "IdP"
	param.MaxIal = 3.0
	param.MaxAal = 3.0

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRegisterNodeAS(t *testing.T) {
	asKey := getPrivateKeyFromString(asPrivK)
	asPublicKeyBytes, err := generatePublicKey(&asKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	asKey2 := getPrivateKeyFromString(allMasterKey)
	asPublicKeyBytes2, err := generatePublicKey(&asKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param did.RegisterNode
	param.NodeName = "AS1"
	param.NodeID = "AS1"
	param.PublicKey = string(asPublicKeyBytes)
	param.MasterPublicKey = string(asPublicKeyBytes2)
	param.Role = "AS"

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodePublicKeyRP(t *testing.T) {
	fnName := "GetNodePublicKey"
	var param = did.GetNodePublicKeyParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodePublicKeyResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	rpKey := getPrivateKeyFromString(rpPrivK)
	rpPublicKeyBytes, err := generatePublicKey(&rpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	expected := string(rpPublicKeyBytes)
	if actual := res.PublicKey; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeMasterPublicKeyRP(t *testing.T) {
	fnName := "GetNodeMasterPublicKey"
	var param = did.GetNodePublicKeyParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodeMasterPublicKeyResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	rpKey := getPrivateKeyFromString(allMasterKey)
	rpPublicKeyBytes, err := generatePublicKey(&rpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	expected := string(rpPublicKeyBytes)
	if actual := res.MasterPublicKey; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodePublicKeyIdP(t *testing.T) {
	fnName := "GetNodePublicKey"
	var param = did.GetNodePublicKeyParam{
		"IdP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodePublicKeyResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	expected := string(idpPublicKeyBytes)
	if actual := res.PublicKey; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodePublicKeyAS(t *testing.T) {
	fnName := "GetNodePublicKey"
	var param = did.GetNodePublicKeyParam{
		"AS1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodePublicKeyResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	asKey := getPrivateKeyFromString(asPrivK)
	asPublicKeyBytes, err := generatePublicKey(&asKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	expected := string(asPublicKeyBytes)
	if actual := res.PublicKey; actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddNodeTokenRP(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.AddNodeTokenParam{
		"RP1",
		111.11,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddNodeTokenIdP(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.AddNodeTokenParam{
		"IdP1",
		222.22,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddNodeTokenAS(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.AddNodeTokenParam{
		"AS1",
		333.33,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeTokenRP(t *testing.T) {
	fnName := "GetNodeToken"
	var param = did.GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetNodeTokenResult{
		111.11,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestReduceNodeTokenRP(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.ReduceNodeTokenParam{
		"RP1",
		61.11,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "ReduceNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeTokenRPAfterReduce(t *testing.T) {
	fnName := "GetNodeToken"
	var param = did.GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetNodeTokenResult{
		50.0,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestSetNodeTokenRP(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.SetNodeTokenParam{
		"RP1",
		100.0,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SetNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeTokenRPAfterSetToken(t *testing.T) {
	fnName := "GetNodeToken"
	var param = did.GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetNodeTokenResult{
		100.0,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDAddService(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.AddServiceParam{
		"statement",
		"Bank statement",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDDeleteService(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.DeleteServiceParam{
		"statement",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "DeleteService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDAddServiceAgain(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.AddServiceParam{
		"statement",
		"Bank statement",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPRegisterMsqDestination(t *testing.T) {

	h := sha256.New()
	h.Write([]byte(userNamespace + userID))
	userHash := h.Sum(nil)

	var users []did.User
	var user = did.User{
		hex.EncodeToString(userHash),
		3,
		true,
	}
	users = append(users, user)

	var param = did.RegisterMsqDestinationParam{
		users,
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterMsqDestination"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetIdpNodes(t *testing.T) {
	fnName := "GetIdpNodes"
	h := sha256.New()
	h.Write([]byte(userNamespace + userID))
	userHash := h.Sum(nil)
	var param = did.GetIdpNodesParam{
		hex.EncodeToString(userHash),
		3,
		3,
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetIdpNodesResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = []did.MsqDestinationNode{
		{
			"IdP1",
			"IdP Number 1 from ...",
			3.0,
			3.0,
		},
	}
	if actual := res.Node; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPRegisterMsqAddress(t *testing.T) {

	var param = did.RegisterMsqAddressParam{
		"IdP1",
		"192.168.3.99",
		8000,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterMsqAddress"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetMsqAddress(t *testing.T) {
	fnName := "GetMsqAddress"
	var param = did.GetMsqAddressParam{
		"IdP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.MsqAddress
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.MsqAddress{
		"192.168.3.99",
		8000,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestASRegisterServiceDestination(t *testing.T) {
	var param = did.RegisterServiceDestinationParam{
		"statement",
		1.1,
		1.2,
	}

	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte("AS1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterServiceDestination"
	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestASRegisterServiceDestination2(t *testing.T) {
	var param = did.RegisterServiceDestinationParam{
		"statement",
		1.1,
		1.2,
	}

	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte("AS1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterServiceDestination"
	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "Duplicate node ID"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDUpdateService(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.UpdateServiceParam{
		"statement",
		"Bank statement (ย้อนหลัง 3 เดือน)",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateService"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetServiceDetail(t *testing.T) {
	fnName := "GetServiceDetail"
	var param = did.GetServiceDetailParam{
		"statement",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.Service
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.Service{
		"Bank statement (ย้อนหลัง 3 เดือน)",
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestASUpdateServiceDestination(t *testing.T) {
	var param = did.UpdateServiceDestinationParam{
		"statement",
		1.4,
		1.5,
	}

	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte("AS1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateServiceDestination"
	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetAsNodesByServiceId(t *testing.T) {
	fnName := "GetAsNodesByServiceId"
	var param = did.GetAsNodesByServiceIdParam{
		"statement",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res did.GetAsNodesByServiceIdResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = `{"node":[{"node_id":"AS1","node_name":"AS1","min_ial":1.4,"min_aal":1.5}]}`
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRPCreateRequest(t *testing.T) {
	var datas []did.DataRequest
	var data1 did.DataRequest
	data1.ServiceID = "statement"
	// data1.As = []string{
	// 	"AS1",
	// 	"AS2",
	// }
	data1.Count = 1
	data1.RequestParamsHash = "hash"

	datas = append(datas, data1)

	var param did.Request
	param.RequestID = "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3

	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte("RP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CreateRequest"
	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeTokenRPAfterCreatRequest(t *testing.T) {
	fnName := "GetNodeToken"
	var param = did.GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetNodeTokenResult{
		99.0,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPDeclareIdentityProof(t *testing.T) {
	var param did.DeclareIdentityProofParam
	param.RequestID = "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
	param.IdentityProof = "Magic"

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "DeclareIdentityProof"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetIdentityProof(t *testing.T) {
	fnName := "GetIdentityProof"
	var param = did.GetIdentityProofParam{
		"IdP1",
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetIdentityProofResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetIdentityProofResult{
		"Magic",
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPCreateIdpResponse(t *testing.T) {
	var param = did.CreateIdpResponseParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		3,
		3,
		"accept",
		"signature",
		"Magic",
		"Magic",
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CreateIdpResponse"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestASSignData(t *testing.T) {
	var param = did.SignDataParam{
		"statement",
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		"sign(data,asKey)",
	}

	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte("AS1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SignData"
	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestASSignData2(t *testing.T) {
	var param = did.SignDataParam{
		"statement",
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		"sign(data,asKey)",
	}

	asKey := getPrivateKeyFromString(asPrivK)
	asNodeID := []byte("AS1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SignData"
	signature, err := rsa.SignPKCS1v15(rand.Reader, asKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, asNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "Duplicate AS ID in answered AS list"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRPSetDataReceived(t *testing.T) {

	var param = did.SetDataReceivedParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		"statement",
		"AS1",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte("RP1")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SetDataReceived"
	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRPSetDataReceived2(t *testing.T) {

	var param = did.SetDataReceivedParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		"statement",
		"AS1",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte("RP1")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SetDataReceived"
	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "Duplicate AS ID in data request"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPCreateRequestSpecial(t *testing.T) {
	var datas []did.DataRequest
	var param did.Request
	param.RequestID = "ef6f4c9c-818b-42b8-8904-3d97c4c55555"
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CreateRequest"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPDeclareIdentityProof2(t *testing.T) {
	var param did.DeclareIdentityProofParam
	param.RequestID = "ef6f4c9c-818b-42b8-8904-3d97c4c55555"
	param.IdentityProof = "Magic"

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "DeclareIdentityProof"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}
func TestIdPCreateIdpResponseForSpecialRequest(t *testing.T) {
	var param = did.CreateIdpResponseParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c55555",
		3,
		3,
		"accept",
		"signature",
		"Magic",
		"Magic",
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CreateIdpResponse"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDSetPrice(t *testing.T) {

	var param = did.SetPriceFuncParam{
		"CreateRequest",
		9.99,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SetPriceFunc"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDGetPrice(t *testing.T) {
	fnName := "GetPriceFunc"
	var param = did.GetPriceFuncParam{
		"CreateRequest",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetPriceFuncResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetPriceFuncResult{
		9.99,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

type Report struct {
	Method string  `json:"method"`
	Price  float64 `json:"price"`
	Data   string  `json:"data"`
}

type GetUsedTokenReportParam struct {
	NodeID string `json:"node_id"`
}

func TestReportGetUsedTokenRP(t *testing.T) {
	fnName := "GetUsedTokenReport"
	var param = GetUsedTokenReportParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res []Report
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	expectedString := `[{"method":"CreateRequest","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6"},{"method":"SetDataReceived","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6"}]`
	var expected []Report
	json.Unmarshal([]byte(expectedString), &expected)

	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestReportGetUsedTokenIdP(t *testing.T) {
	fnName := "GetUsedTokenReport"
	var param = GetUsedTokenReportParam{
		"IdP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res []Report
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	// fmt.Println(string(resultString))
	expectedString := `[{"method":"RegisterMsqDestination","price":1,"data":""},{"method":"RegisterMsqAddress","price":1,"data":""},{"method":"DeclareIdentityProof","price":1,"data":""},{"method":"CreateIdpResponse","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6"},{"method":"CreateRequest","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c55555"},{"method":"DeclareIdentityProof","price":1,"data":""},{"method":"CreateIdpResponse","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c55555"}]`
	var expected []Report
	json.Unmarshal([]byte(expectedString), &expected)

	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestReportGetUsedTokenAS(t *testing.T) {
	fnName := "GetUsedTokenReport"
	var param = GetUsedTokenReportParam{
		"AS1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res []Report
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}

	expectedString := `[{"method":"RegisterServiceDestination","price":1,"data":""},{"method":"UpdateServiceDestination","price":1,"data":""},{"method":"SignData","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6"}]`
	var expected []Report
	json.Unmarshal([]byte(expectedString), &expected)

	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetRequestDetail1(t *testing.T) {
	fnName := "GetRequestDetail"
	var param = did.GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	// fmt.Println(string(resultString))
	var expected = `{"request_id":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6","min_idp":1,"min_aal":3,"min_ial":3,"request_timeout":259200,"data_request_list":[{"service_id":"statement","as_id_list":[],"min_as":1,"request_params_hash":"hash","answered_as_id_list":["AS1"],"received_data_from_list":["AS1"]}],"request_message_hash":"hash('Please allow...')","response_list":[{"ial":3,"aal":3,"status":"accept","signature":"signature","identity_proof":"Magic","private_proof_hash":"Magic","idp_id":"IdP1","valid_proof":null,"valid_ial":null}],"closed":false,"timed_out":false,"special":false,"mode":3}`
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRPCloseRequest(t *testing.T) {

	var res []did.ResponseValid
	var res1 did.ResponseValid
	res1.IdpID = "IdP1"
	res1.ValidIal = true
	res1.ValidProof = true
	res = append(res, res1)
	var param = did.CloseRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		res,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte("RP1")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CloseRequest"
	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetRequestClosed(t *testing.T) {
	fnName := "GetRequest"
	var param = did.GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetRequestResult{
		true,
		false,
		"hash('Please allow...')",
		3,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetRequestDetail2(t *testing.T) {
	fnName := "GetRequestDetail"
	var param = did.GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	// fmt.Println(string(resultString))
	var expected = `{"request_id":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6","min_idp":1,"min_aal":3,"min_ial":3,"request_timeout":259200,"data_request_list":[{"service_id":"statement","as_id_list":[],"min_as":1,"request_params_hash":"hash","answered_as_id_list":["AS1"],"received_data_from_list":["AS1"]}],"request_message_hash":"hash('Please allow...')","response_list":[{"ial":3,"aal":3,"status":"accept","signature":"signature","identity_proof":"Magic","private_proof_hash":"Magic","idp_id":"IdP1","valid_proof":true,"valid_ial":true}],"closed":true,"timed_out":false,"special":false,"mode":3}`
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestCreateRequest(t *testing.T) {
	var datas []did.DataRequest
	var data1 did.DataRequest
	data1.ServiceID = "statement"
	data1.As = []string{
		"AS1",
		"AS2",
	}
	data1.Count = 2
	data1.RequestParamsHash = "hash"

	var data2 did.DataRequest
	data2.ServiceID = "credit"
	data2.As = []string{
		"AS1",
		"AS2",
	}
	data2.Count = 2
	data2.RequestParamsHash = "hash"

	datas = append(datas, data1)
	datas = append(datas, data2)

	var param did.Request
	param.RequestID = "ef6f4c9c-818b-42b8-8904-3d97c4c11111"
	param.MinIdp = 1
	param.MinIal = 3
	param.MinAal = 3
	param.Timeout = 259200
	param.DataRequestList = datas
	param.MessageHash = "hash('Please allow...')"
	param.Mode = 3

	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte("RP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CreateRequest"
	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPDeclareIdentityProof3(t *testing.T) {
	var param did.DeclareIdentityProofParam
	param.RequestID = "ef6f4c9c-818b-42b8-8904-3d97c4c11111"
	param.IdentityProof = "Magic"

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "DeclareIdentityProof"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPCreateIdpResponse2(t *testing.T) {
	var param = did.CreateIdpResponseParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c11111",
		3,
		3,
		"accept",
		"signature",
		"Magic",
		"Magic",
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CreateIdpResponse"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRPTimeOutRequest(t *testing.T) {

	var res []did.ResponseValid
	var res1 did.ResponseValid
	res1.IdpID = "IdP1"
	res1.ValidIal = false
	res1.ValidProof = false
	res = append(res, res1)
	var param = did.TimeOutRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c11111",
		res,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	rpKey := getPrivateKeyFromString(rpPrivK)
	rpNodeID := []byte("RP1")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "TimeOutRequest"
	signature, err := rsa.SignPKCS1v15(rand.Reader, rpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, rpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetRequestDetail3(t *testing.T) {
	fnName := "GetRequestDetail"
	var param = did.GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c11111",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	// fmt.Println(string(resultString))
	var expected = `{"request_id":"ef6f4c9c-818b-42b8-8904-3d97c4c11111","min_idp":1,"min_aal":3,"min_ial":3,"request_timeout":259200,"data_request_list":[{"service_id":"statement","as_id_list":["AS1","AS2"],"min_as":2,"request_params_hash":"hash","answered_as_id_list":[],"received_data_from_list":[]},{"service_id":"credit","as_id_list":["AS1","AS2"],"min_as":2,"request_params_hash":"hash","answered_as_id_list":[],"received_data_from_list":[]}],"request_message_hash":"hash('Please allow...')","response_list":[{"ial":3,"aal":3,"status":"accept","signature":"signature","identity_proof":"Magic","private_proof_hash":"Magic","idp_id":"IdP1","valid_proof":false,"valid_ial":false}],"closed":false,"timed_out":true,"special":false,"mode":3}`
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetRequestTimedOut(t *testing.T) {
	fnName := "GetRequest"
	var param = did.GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c11111",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = did.GetRequestResult{
		false,
		true,
		"hash('Please allow...')",
		3,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddNamespaceCID(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	nodeID := "NDID"

	var funcparam did.Namespace
	funcparam.Namespace = "CID"
	funcparam.Description = "Citizen ID"

	funcparamJSON, err := json.Marshal(funcparam)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(funcparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddNamespace"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), funcparamJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddNamespaceTel(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	nodeID := "NDID"

	var funcparam did.Namespace
	funcparam.Namespace = "Tel"
	funcparam.Description = "Tel number"

	funcparamJSON, err := json.Marshal(funcparam)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(funcparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddNamespace"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), funcparamJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestDeleteNamespace(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	nodeID := "NDID"

	var funcparam did.DeleteNamespaceParam
	funcparam.Namespace = "Tel"

	funcparamJSON, err := json.Marshal(funcparam)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(funcparamJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "DeleteNamespace"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), funcparamJSON, []byte(nonce), signature, []byte(nodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNamespaceList(t *testing.T) {
	fnName := "GetNamespaceList"
	paramJSON := []byte("")
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res []did.Namespace
	err := json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = []did.Namespace{
		did.Namespace{
			"CID",
			"Citizen ID",
		},
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPCreateIdentity(t *testing.T) {

	var param = did.CreateIdentityParam{
		"accessor_id",
		"accessor_type",
		"accessor_public_key",
		"accessor_group_id",
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "CreateIdentity"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPAddAccessorMethod(t *testing.T) {

	var param = did.AccessorMethod{
		"accessor_id_2",
		"accessor_type_2",
		"accessor_public_key_2",
		"accessor_group_id",
		"ef6f4c9c-818b-42b8-8904-3d97c4c55555",
	}

	idpKey := getPrivateKeyFromString(idpPrivK)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "AddAccessorMethod"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

type CheckExistingIdentityParam struct {
	HashID string `json:"hash_id"`
}

func TestQueryCheckExistingIdentity(t *testing.T) {
	fnName := "CheckExistingIdentity"
	h := sha256.New()
	h.Write([]byte(userNamespace + userID))
	userHash := h.Sum(nil)
	var param = CheckExistingIdentityParam{
		hex.EncodeToString(userHash),
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var expected = `{"exist":true}`
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

type GetAccessorGroupIDParam struct {
	AccessorID string `json:"accessor_id"`
}

func TestQueryGetAccessorGroupID(t *testing.T) {
	fnName := "GetAccessorGroupID"
	var param = GetAccessorGroupIDParam{
		"accessor_id_2",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var expected = `{"accessor_group_id":"accessor_group_id"}`
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

type GetAccessorKeyParam struct {
	AccessorID string `json:"accessor_id"`
}

func TestQueryGetAccessorKey(t *testing.T) {
	fnName := "GetAccessorKey"
	var param = GetAccessorKeyParam{
		"accessor_id",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var expected = `{"accessor_public_key":"accessor_public_key"}`
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRegisterNodeIDP2(t *testing.T) {
	idpKey := getPrivateKeyFromString(idpPrivK3)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param did.RegisterNode
	param.NodeID = "IdP2"
	param.PublicKey = string(idpPublicKeyBytes)
	param.Role = "IdP"
	param.MaxIal = 3.0
	param.MaxAal = 3.0

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetIdpNodes2(t *testing.T) {
	fnName := "GetIdpNodes"
	var param did.GetIdpNodesParam
	param.MinIal = 3
	param.MinAal = 3

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res did.GetIdpNodesResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = []did.MsqDestinationNode{
		{
			"IdP1",
			"IdP Number 1 from ...",
			3.0,
			3.0,
		},
		{
			"IdP2",
			"",
			3.0,
			3.0,
		},
	}
	if actual := res.Node; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPUpdateNode(t *testing.T) {

	masterIdPKey := getPrivateKeyFromString(allMasterKey)

	idpKey2 := getPrivateKeyFromString(idpPrivK2)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param = did.UpdateNodeParam{
		string(idpPublicKeyBytes2),
		"",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	idpNodeID := []byte("IdP1")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, masterIdPKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

type SetValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}

func TestSetValidator(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param SetValidatorParam
	param.PublicKey = `7/ThXSVOL7YkcpcJ8iatM+EXOlXv8aFtpsVAmWwMdC4=`
	// param.PublicKey = `5/6rEo7aQYq31J32higcxi3i8xp9MG/r5Ho5NemwZ+g=`
	param.Power = 20

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SetValidator"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetServiceList(t *testing.T) {
	fnName := "GetServiceList"
	paramJSON := []byte("")
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res []did.ServiceDetail
	err := json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = []did.ServiceDetail{
		did.ServiceDetail{
			"statement",
			"Bank statement (ย้อนหลัง 3 เดือน)",
		},
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestUpdateNodeByNDID(t *testing.T) {
	var param did.UpdateNodeByNDIDParam
	param.NodeID = "IdP1"
	param.MaxIal = 2.3
	param.MaxAal = 2.4

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateNodeByNDID"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeInfo(t *testing.T) {
	fnName := "GetNodeInfo"
	var param did.GetNodeInfoParam
	param.NodeID = "IdP1"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 1 from ...","role":"IdP","max_ial":2.3,"max_aal":2.4}`)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryCheckExistingAccessorID(t *testing.T) {
	fnName := "CheckExistingAccessorID"
	var param did.CheckExistingAccessorIDParam
	param.AccessorID = "accessor_id"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	expected := string(`{"exist":true}`)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryCheckExistingAccessorGroupID(t *testing.T) {
	fnName := "CheckExistingAccessorGroupID"
	var param did.CheckExistingAccessorGroupIDParam
	param.AccessorGroupID = "accessor_group_id"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	expected := string(`{"exist":true}`)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPUpdateIdentity(t *testing.T) {

	h := sha256.New()
	h.Write([]byte(userNamespace + userID))
	userHash := h.Sum(nil)

	var param did.UpdateIdentityParam
	param.HashID = hex.EncodeToString(userHash)
	param.Ial = 2.2

	idpKey := getPrivateKeyFromString(idpPrivK2)
	idpNodeID := []byte("IdP1")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateIdentity"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetIdentityInfo(t *testing.T) {
	fnName := "GetIdentityInfo"
	var param did.GetIdentityInfoParam
	h := sha256.New()
	h.Write([]byte(userNamespace + userID))
	userHash := h.Sum(nil)
	param.NodeID = "IdP1"
	param.HashID = hex.EncodeToString(userHash)
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	expected := string(`{"ial":2.2}`)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetDataSignature(t *testing.T) {
	fnName := "GetDataSignature"
	var param did.GetDataSignatureParam
	param.NodeID = "AS1"
	param.RequestID = "ef6f4c9c-818b-42b8-8904-3d97c4c520f6"
	param.ServiceID = "statement"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	expected := string(`{"signature":"sign(data,asKey)"}`)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

// TODO add more test about DPKI

func TestRegisterNodeIDP4(t *testing.T) {
	idpKey := getPrivateKeyFromString(idpPrivK4)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	idpKey2 := getPrivateKeyFromString(allMasterKey)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param did.RegisterNode
	param.NodeID = "IdP4"
	param.PublicKey = string(idpPublicKeyBytes)
	param.MasterPublicKey = string(idpPublicKeyBytes2)
	param.NodeName = "IdP Number 4 from ..."
	param.Role = "IdP"
	param.MaxIal = 3.0
	param.MaxAal = 3.0

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRegisterNodeIDP5(t *testing.T) {
	idpKey := getPrivateKeyFromString(idpPrivK5)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	idpKey2 := getPrivateKeyFromString(allMasterKey)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param did.RegisterNode
	param.NodeID = "IdP5"
	param.PublicKey = string(idpPublicKeyBytes)
	param.MasterPublicKey = string(idpPublicKeyBytes2)
	param.NodeName = "IdP Number 5 from ..."
	param.Role = "IdP"
	param.MaxIal = 3.0
	param.MaxAal = 3.0

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := []byte("NDID")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, ndidNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestSetNodeTokenIDP4(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.SetNodeTokenParam{
		"IdP4",
		100.0,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SetNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestSetNodeTokenIDP5(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = did.SetNodeTokenParam{
		"IdP5",
		100.0,
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "SetNodeToken"
	signature, err := rsa.SignPKCS1v15(rand.Reader, ndidKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, []byte(ndidNodeID))
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPUpdateNode4(t *testing.T) {

	masterIdPKey := getPrivateKeyFromString(allMasterKey)

	idpKey2 := getPrivateKeyFromString(idpPrivK5)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param = did.UpdateNodeParam{
		string(idpPublicKeyBytes2),
		"",
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	idpNodeID := []byte("IdP4")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, masterIdPKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPUpdateNode5(t *testing.T) {

	masterIdPKey := getPrivateKeyFromString(allMasterKey)

	idpKey2 := getPrivateKeyFromString(idpPrivK4)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param = did.UpdateNodeParam{
		string(idpPublicKeyBytes2),
		string(idpPublicKeyBytes2),
	}

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	idpNodeID := []byte("IdP5")

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "UpdateNode"
	signature, err := rsa.SignPKCS1v15(rand.Reader, masterIdPKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeInfoIdP4(t *testing.T) {
	fnName := "GetNodeInfo"
	var param did.GetNodeInfoParam
	param.NodeID = "IdP4"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu9+CK/vznpXtAUC0QhuJ\ngYKCfMMBiIgVcp2A+e+SsKvv6ESQ72R8K6nQAhH2MGtnj3ScLI0tMwCtgotWCEGi\nyUXKXLVTiqAqtwflCUVuxCDVuvOm3GQCxvwzE34jEgbGZ33G3tV7uKTtifhoJzVY\nD+WkZVslBhaBgQCUewCX4zkCCTYC5VEhkr7K8HGEr6n1eBOO5VORCkrHKYoZK7eu\nNjyWvWYyVN07F8K0RhgIF9Xsa6Tiu1Yf8zuyJ/awR6U4Nw+oTkvRpx64+caBNYgR\n4n8peg9ZJeTAwV49o1ymx34pPjHUgSdpyhZX4i3z9ji+o7KbNkA/O0l+3doMuH1e\nxwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 4 from ...","role":"IdP","max_ial":3,"max_aal":3}`)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetNodeInfoIdP5(t *testing.T) {
	fnName := "GetNodeInfo"
	var param did.GetNodeInfoParam
	param.NodeID = "IdP5"
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	expected := string(`{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApbxaA5aKnkpnV7+dMW5x\n7iEINouvjhQ8gl6+8A6ApiVbYIzJCCaexU9mn7jDP634SyjFNSxzhjklEm7qFPaH\nOk1FfX6tk5i5uGWifRQHueXhXjR8HSBkjQAoZ0eqBqTsxsSpASsT4qoBKtsIVN7X\nHdh9Mqz+XAkq4T6vtdaocduarNG6ALZFkX+pAgkCj4hIhRmHjlyYIh1yOZw1KM3T\nHkM9noP2AYEH2MBHCzuu+bifCwurOBq+ZKAdfroCG4rPGfOXuDQK8BHpru1lg0jd\nAmbbqMyGpAsF+WjW4V2rcTMFZOoYFYE5m2ssxC4O9h3f/H2gBtjjWzYv6bRC6ZdP\n2wIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 5 from ...","role":"IdP","max_ial":3,"max_aal":3}`)
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdP4RegisterMsqDestination(t *testing.T) {

	h := sha256.New()
	h.Write([]byte(userNamespace + userID))
	userHash := h.Sum(nil)

	var users []did.User
	var user = did.User{
		hex.EncodeToString(userHash),
		3,
		false,
	}
	users = append(users, user)

	var param = did.RegisterMsqDestinationParam{
		users,
	}

	idpKey := getPrivateKeyFromString(idpPrivK5)
	idpNodeID := []byte("IdP4")

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}

	nonce := base64.StdEncoding.EncodeToString([]byte(common.RandStr(12)))
	PSSmessage := append(paramJSON, []byte(nonce)...)
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)

	fnName := "RegisterMsqDestination"
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}
