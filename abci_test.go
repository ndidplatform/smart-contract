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

	"github.com/tendermint/tmlibs/common"
)

type InitNDID struct {
	NodeID    string `json:"node_id"`
	PublicKey string `json:"public_key"`
}

type RegisterNode struct {
	NodeID          string  `json:"node_id"`
	PublicKey       string  `json:"public_key"`
	MasterPublicKey string  `json:"master_public_key"`
	NodeName        string  `json:"node_name"`
	Role            string  `json:"role"`
	MaxIal          float64 `json:"max_ial"`
	MaxAal          float64 `json:"max_aal"`
}

type DataRequest struct {
	ServiceID         string   `json:"service_id"`
	As                []string `json:"as_id_list"`
	Count             int      `json:"count"`
	RequestParamsHash string   `json:"request_params_hash"`
}

type Request struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          int           `json:"min_aal"`
	MinIal          int           `json:"min_ial"`
	Timeout         int           `json:"request_timeout"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"request_message_hash"`
	Special         bool          `json:"special"`
}

type User struct {
	HashID string `json:"hash_id"`
	Ial    int    `json:"ial"`
}

type RegisterMsqDestination struct {
	Users  []User `json:"users"`
	NodeID string `json:"node_id"`
}

type Response struct {
	RequestID     string `json:"request_id"`
	Aal           int    `json:"aal"`
	Ial           int    `json:"ial"`
	Status        string `json:"status"`
	Signature     string `json:"signature"`
	IdentityProof string `json:"identity_proof"`
}

type SignDataParam struct {
	ServiceID string `json:"service_id"`
	RequestID string `json:"request_id"`
	Signature string `json:"signature"`
}

type ResponseTx struct {
	Result struct {
		Height  int `json:"height"`
		CheckTx struct {
			Code int      `json:"code"`
			Log  string   `json:"log"`
			Fee  struct{} `json:"fee"`
		} `json:"check_tx"`
		DeliverTx struct {
			Log string   `json:"log"`
			Fee struct{} `json:"fee"`
		} `json:"deliver_tx"`
		Hash string `json:"hash"`
	} `json:"result"`
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
}

type GetNodePublicKey struct {
	NodeID string `json:"node_id"`
}

type GetNodePublicKeyResult struct {
	PublicKey string `json:"public_key"`
}

type GetIdpNodesParam struct {
	HashID string  `json:"hash_id"`
	MinIal float64 `json:"min_ial"`
	MinAal float64 `json:"min_aal"`
}

type MsqDestinationNode struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	MaxIal float64 `json:"max_ial"`
	MaxAal float64 `json:"max_aal"`
}

type GetIdpNodesResult struct {
	Node []MsqDestinationNode `json:"node"`
}

type GetAccessorMethodParam struct {
	AccessorID string `json:"accessor_id"`
}

type GetAccessorMethodResult struct {
	AccessorType string `json:"accessor_type"`
	AccessorKey  string `json:"accessor_key"`
	Commitment   string `json:"commitment"`
}

type GetRequestParam struct {
	RequestID string `json:"requestId"`
}

type CloseRequestParam struct {
	RequestID string `json:"requestId"`
}

type TimeOutRequestParam struct {
	RequestID string `json:"requestId"`
}

type GetRequestResult struct {
	Status      string `json:"status"`
	IsClosed    bool   `json:"closed"`
	IsTimedOut  bool   `json:"timed_out"`
	MessageHash string `json:"request_message_hash"`
}

type RegisterServiceDestinationParam struct {
	AsServiceID string  `json:"service_id"`
	NodeID      string  `json:"node_id"`
	MinIal      float64 `json:"min_ial"`
	MinAal      float64 `json:"min_aal"`
}

type Service struct {
	ServiceName string `json:"service_name"`
}

type GetServiceDetailParam struct {
	AsServiceID string `json:"service_id"`
}

type GetAsNodesByServiceIdParam struct {
	AsServiceID string `json:"service_id"`
}

type ASNode struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	MinIal      float64 `json:"min_ial"`
	MinAal      float64 `json:"min_aal"`
	ServiceName string  `json:"service_name"`
}

type GetAsNodesByServiceIdResult struct {
	Node []ASNode `json:"node"`
}

type RegisterMsqAddressParam struct {
	NodeID string `json:"node_id"`
	IP     string `json:"ip"`
	Port   int64  `json:"port"`
}

type GetMsqAddressParam struct {
	NodeID string `json:"node_id"`
}

type MsqAddress struct {
	IP   string `json:"ip"`
	Port int64  `json:"port"`
}

type RequestDetailResult struct {
	RequestID       string        `json:"request_id"`
	MinIdp          int           `json:"min_idp"`
	MinAal          int           `json:"min_aal"`
	MinIal          int           `json:"min_ial"`
	Timeout         int           `json:"request_timeout"`
	DataRequestList []DataRequest `json:"data_request_list"`
	MessageHash     string        `json:"request_message_hash"`
	Responses       []Response    `json:"responses"`
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

type SetNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

type AddNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

type ReduceNodeTokenParam struct {
	NodeID string  `json:"node_id"`
	Amount float64 `json:"amount"`
}

type GetNodeTokenParam struct {
	NodeID string `json:"node_id"`
}

type GetNodeTokenResult struct {
	Amount float64 `json:"amount"`
}

type SetPriceFuncParam struct {
	Func  string  `json:"func"`
	Price float64 `json:"price"`
}

type GetPriceFuncParam struct {
	Func string `json:"func"`
}

type GetPriceFuncResult struct {
	Price float64 `json:"price"`
}

type Namespace struct {
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
}

type DeleteNamespaceParam struct {
	Namespace string `json:"namespace"`
}

type UpdateNodeParam struct {
	PublicKey       string `json:"public_key"`
	MasterPublicKey string `json:"master_public_key"`
}

type CreateIdentityParam struct {
	AccessorID        string `json:"accessor_id"`
	AccessorType      string `json:"accessor_type"`
	AccessorPublicKey string `json:"accessor_public_key"`
	AccessorGroupID   string `json:"accessor_group_id"`
}

type AccessorMethod struct {
	AccessorID        string `json:"accessor_id"`
	AccessorType      string `json:"accessor_type"`
	AccessorPublicKey string `json:"accessor_public_key"`
	AccessorGroupID   string `json:"accessor_group_id"`
	RequestID         string `json:"request_id"`
}

type RegisterServiceParam struct {
	AsServiceID string `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type DeleteServiceParam struct {
	AsServiceID string `json:"service_id"`
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

func callTendermint(fnName []byte, param []byte, nonce []byte, signature []byte, publicKey []byte) (interface{}, error) {
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
	path = append(path, publicKey...)

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

var userNamespace = "cid"
var userID = "1234567890123"

func TestInitNDID(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidpublicKeyBytes, err := generatePublicKey(&ndidKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}
	var initNDIDparam InitNDID
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

	rpKey2 := getPrivateKeyFromString(rpPrivK2)
	rpPublicKeyBytes2, err := generatePublicKey(&rpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param RegisterNode
	param.NodeID = "RP1"
	param.PublicKey = string(rpPublicKeyBytes)
	param.MasterPublicKey = string(rpPublicKeyBytes2)
	param.Role = "RP"

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

func TestRegisterNodeIDP(t *testing.T) {
	idpKey := getPrivateKeyFromString(idpPrivK)
	idpPublicKeyBytes, err := generatePublicKey(&idpKey.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	idpKey2 := getPrivateKeyFromString(idpPrivK2)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param RegisterNode
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

	asKey2 := getPrivateKeyFromString(asPrivK2)
	asPublicKeyBytes2, err := generatePublicKey(&asKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param RegisterNode
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
	var param = GetNodePublicKey{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetNodePublicKeyResult
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

func TestQueryGetNodePublicKeyIdP(t *testing.T) {
	fnName := "GetNodePublicKey"
	var param = GetNodePublicKey{
		"IdP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetNodePublicKeyResult
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
	var param = GetNodePublicKey{
		"AS1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetNodePublicKeyResult
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

	var param = AddNodeTokenParam{
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

	var param = AddNodeTokenParam{
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

	var param = AddNodeTokenParam{
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
	var param = GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetNodeTokenResult{
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

	var param = ReduceNodeTokenParam{
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
	var param = GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetNodeTokenResult{
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

	var param = SetNodeTokenParam{
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
	var param = GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetNodeTokenResult{
		100.0,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDRegisterService(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = RegisterServiceParam{
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

	fnName := "RegisterService"
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

	var param = DeleteServiceParam{
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

func TestNDIDRegisterServiceAgain(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param = RegisterServiceParam{
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

	fnName := "RegisterService"
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

	var users []User
	var user = User{
		hex.EncodeToString(userHash),
		3,
	}
	users = append(users, user)

	var param = RegisterMsqDestination{
		users,
		"IdP1",
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
	var param = GetIdpNodesParam{
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

	var res GetIdpNodesResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = []MsqDestinationNode{
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

	var param = RegisterMsqAddressParam{
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
	var param = GetMsqAddressParam{
		"IdP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res MsqAddress
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = MsqAddress{
		"192.168.3.99",
		8000,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestASRegisterServiceDestination(t *testing.T) {
	var param = RegisterServiceDestinationParam{
		"statement",
		"AS1",
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

func TestQueryGetServiceDetail(t *testing.T) {
	fnName := "GetServiceDetail"
	var param = GetServiceDetailParam{
		"statement",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res Service
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = Service{
		"Bank statement",
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetAsNodesByServiceId(t *testing.T) {
	fnName := "GetAsNodesByServiceId"
	var param = GetAsNodesByServiceIdParam{
		"statement",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		log.Fatal(err.Error())
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var res GetAsNodesByServiceIdResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = `{"node":[{"id":"AS1","name":"AS1","min_ial":1.1,"min_aal":1.2,"service_name":"Bank statement"}]}`
	if actual := string(resultString); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRPCreateRequest(t *testing.T) {
	var datas []DataRequest
	var data1 DataRequest
	data1.ServiceID = "statement"
	data1.As = []string{
		"AS1",
		"AS2",
	}
	data1.Count = 1
	data1.RequestParamsHash = "hash"

	datas = append(datas, data1)

	var param = Request{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		1,
		3,
		3,
		259200,
		datas,
		"hash('Please allow...')",
		true,
	}

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
	var param = GetNodeTokenParam{
		"RP1",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetNodeTokenResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetNodeTokenResult{
		99.0,
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetRequestPending(t *testing.T) {
	fnName := "GetRequest"
	var param = GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetRequestResult{
		"pending",
		false,
		false,
		"hash('Please allow...')",
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestIdPCreateIdpResponse(t *testing.T) {
	var param = Response{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
		3,
		3,
		"accept",
		"signature",
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
	var param = SignDataParam{
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

func TestQueryGetRequestComplete(t *testing.T) {
	fnName := "GetRequest"
	var param = GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetRequestResult{
		"completed",
		false,
		false,
		"hash('Please allow...')",
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestQueryGetRequestDetail(t *testing.T) {
	fnName := "GetRequestDetail"
	var param = GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)
	var expected = `{"request_id":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6","min_idp":1,"min_aal":3,"min_ial":3,"request_timeout":259200,"data_request_list":[{"service_id":"statement","as_id_list":["AS1","AS2"],"count":1,"request_params_hash":"hash","answered_as_id_list":["AS1"]}],"request_message_hash":"hash('Please allow...')","responses":[{"request_id":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6","aal":3,"ial":3,"status":"accept","signature":"signature","identity_proof":"Magic","private_proof_hash":"","idp_id":"IdP1"}],"closed":false,"timed_out":false,"special":true,"status":"completed"}`
	if actual := string(resultString); actual != expected {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestNDIDSetPrice(t *testing.T) {

	var param = SetPriceFuncParam{
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
	var param = GetPriceFuncParam{
		"CreateRequest",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetPriceFuncResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetPriceFuncResult{
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

	expectedString := `[{"method":"CreateRequest","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6"}]`
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

	expectedString := `[{"method":"RegisterMsqDestination","price":1,"data":""},{"method":"RegisterMsqAddress","price":1,"data":""},{"method":"CreateIdpResponse","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6"}]`
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

	expectedString := `[{"method":"RegisterServiceDestination","price":1,"data":""},{"method":"SignData","price":1,"data":"ef6f4c9c-818b-42b8-8904-3d97c4c520f6"}]`
	var expected []Report
	json.Unmarshal([]byte(expectedString), &expected)

	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestRPCloseRequest(t *testing.T) {

	var param = CloseRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
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
	var param = GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetRequestResult{
		"completed",
		true,
		false,
		"hash('Please allow...')",
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestCreateRequest(t *testing.T) {
	var datas []DataRequest
	var data1 DataRequest
	data1.ServiceID = "statement"
	data1.As = []string{
		"AS1",
		"AS2",
	}
	data1.Count = 2
	data1.RequestParamsHash = "hash"

	var data2 DataRequest
	data2.ServiceID = "credit"
	data2.As = []string{
		"AS1",
		"AS2",
	}
	data2.Count = 2
	data2.RequestParamsHash = "hash"

	datas = append(datas, data1)
	datas = append(datas, data2)

	var param = Request{
		"ef6f4c9c-818b-42b8-8904-3d97c4c11111",
		1,
		1,
		1,
		259200,
		datas,
		"hash('Please allow...')",
		false,
	}

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

func TestRPTimeOutRequest(t *testing.T) {

	var param = TimeOutRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c11111",
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

func TestQueryGetRequestTimedOut(t *testing.T) {
	fnName := "GetRequest"
	var param = GetRequestParam{
		"ef6f4c9c-818b-42b8-8904-3d97c4c11111",
	}
	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetRequestResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = GetRequestResult{
		"pending",
		false,
		true,
		"hash('Please allow...')",
	}
	if actual := res; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

func TestAddNamespaceCID(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	nodeID := "NDID"

	var funcparam Namespace
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

	var funcparam Namespace
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

	var funcparam DeleteNamespaceParam
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

	var res []Namespace
	err := json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = []Namespace{
		Namespace{
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

	var param = CreateIdentityParam{
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

	var param = AccessorMethod{
		"accessor_id_2",
		"accessor_type_2",
		"accessor_public_key_2",
		"accessor_group_id",
		"ef6f4c9c-818b-42b8-8904-3d97c4c520f6",
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

	var param RegisterNode
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
	var param GetIdpNodesParam
	param.MinIal = 3
	param.MinAal = 3

	paramJSON, err := json.Marshal(param)
	if err != nil {
		fmt.Println("error:", err)
	}
	result, _ := queryTendermint([]byte(fnName), paramJSON)
	resultObj, _ := result.(ResponseQuery)
	resultString, _ := base64.StdEncoding.DecodeString(resultObj.Result.Response.Value)

	var res GetIdpNodesResult
	err = json.Unmarshal(resultString, &res)
	if err != nil {
		log.Fatal(err.Error())
	}
	var expected = []MsqDestinationNode{
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

	idpKey2 := getPrivateKeyFromString(idpPrivK2)
	idpPublicKeyBytes2, err := generatePublicKey(&idpKey2.PublicKey)
	if err != nil {
		log.Fatal(err.Error())
	}

	var param = UpdateNodeParam{
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
	signature, err := rsa.SignPKCS1v15(rand.Reader, idpKey2, newhash, hashed)
	result, _ := callTendermint([]byte(fnName), paramJSON, []byte(nonce), signature, idpNodeID)
	resultObj, _ := result.(ResponseTx)
	expected := "success"
	if actual := resultObj.Result.DeliverTx.Log; actual != expected {
		t.Errorf("\n"+`CheckTx log: "%s"`, resultObj.Result.CheckTx.Log)
		t.Fatalf("FAIL: %s\nExpected: %#v\nActual: %#v", fnName, expected, actual)
	}
	t.Logf("PASS: %s", fnName)
}

type UpdateValidatorParam struct {
	PublicKey string `json:"public_key"`
	Power     int64  `json:"power"`
}

func TestUpdateValidator(t *testing.T) {
	ndidKey := getPrivateKeyFromString(ndidPrivK)
	ndidNodeID := "NDID"

	var param UpdateValidatorParam
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

	fnName := "UpdateValidator"
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

// TODO add more test about DPKI
