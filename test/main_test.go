package test

import (
	"strings"
	"testing"

	"github.com/ndidplatform/smart-contract/v4/test/as"
	"github.com/ndidplatform/smart-contract/v4/test/common"
	"github.com/ndidplatform/smart-contract/v4/test/data"
	"github.com/ndidplatform/smart-contract/v4/test/idp"
	"github.com/ndidplatform/smart-contract/v4/test/ndid"
	"github.com/ndidplatform/smart-contract/v4/test/query"
)

func TestNDIDInitNDID(t *testing.T) {
	ndid.TestInitNDID(t)
}

func TestNDIDSetAndGetAllowedMinIalForRegisterIdentityAtFirstIdp(t *testing.T) {
	ndid.TestNDIDSetAllowedMinIalForRegisterIdentityAtFirstIdp(t)
	ndid.TestQueryGetAllowedMinIalForRegisterIdentityAtFirstIdp(t)
}

func TestNDIDAddNamespace(t *testing.T) {
	ndid.TestAddNamespace(t, data.UserNamespace1)
	ndid.TestAddNamespace(t, data.UserNamespace2)
	ndid.TestAddNamespace(t, data.UserNamespace3)
	ndid.TestQueryGetNamespaceList(t, `[{"namespace":"cid","description":"Citizen ID","active":true,"allowed_identifier_count_in_reference_group":1,"allowed_active_identifier_count_in_reference_group":1},{"namespace":"passport","description":"Passport","active":true},{"namespace":"some_id","description":"Some ID","active":true}]`)
}

func TestNDIDUpdateNamespace(t *testing.T) {
	ndid.TestNDIDUpdateNamespace(t)
	ndid.TestQueryGetNamespaceList(t, `[{"namespace":"cid","description":"Citizen ID","active":true,"allowed_identifier_count_in_reference_group":1,"allowed_active_identifier_count_in_reference_group":1},{"namespace":"passport","description":"Passport","active":true},{"namespace":"some_id","description":"Some ID","active":true,"allowed_identifier_count_in_reference_group":2,"allowed_active_identifier_count_in_reference_group":2}]`)
}

func TestNDIDRegisterNode(t *testing.T) {
	ndid.TestRegisterNode(t, data.IdP1)
	ndid.TestRegisterNode(t, data.IdP2)
	ndid.TestRegisterNode(t, data.IdPAgent1)
	ndid.TestRegisterNode(t, data.AS1)
	ndid.TestRegisterNode(t, data.AS2)
}

func TestNDIDSetNodeToken(t *testing.T) {
	ndid.TestSetNodeToken(t, data.IdP1, 100.0)
	ndid.TestSetNodeToken(t, data.IdP2, 100.0)
	ndid.TestSetNodeToken(t, data.AS1, 100.0)
	ndid.TestSetNodeToken(t, data.AS2, 100.0)
}

func TestNodesSetMqAddresses(t *testing.T) {
	common.TestSetMqAddresses(t, data.IdP1, data.IdpPrivK1, "192.168.3.99", 8000)
	common.TestSetMqAddresses(t, data.IdP2, data.IdpPrivK2, "192.168.3.100", 8000)
	common.TestSetMqAddresses(t, data.AS1, data.AsPrivK1, "192.168.3.102", 8000)
	common.TestSetMqAddresses(t, data.AS2, data.AsPrivK2, "192.168.3.103", 8000)
}

func TestIdP1RegisterIdentity(t *testing.T) {
	query.TestQueryCheckExistingIdentity(t, data.UserNamespace1, data.UserID1, `{"exist":false}`)
	common.TestCreateRequest(t, data.RequestID1.String())
	common.TestCloseRequest(t, data.RequestID1.String())
	idp.TestRegisterIdentity(t, 1, "Please input reference group code")
	idp.TestRegisterIdentity(t, 2, "Identifier count is greater than allowed identifier count")
	idp.TestRegisterIdentity(t, 3, "Namespace is invalid")
	idp.TestRegisterIdentity(t, 4, "success")
	query.TestGetIdentityInfo(t, 1, `{"ial":3,"mode_list":[2]}`)
	query.TestGetIdentityInfo(t, 2, `{"ial":3,"mode_list":[2]}`)
	query.TestQueryCheckExistingIdentity(t, data.UserNamespace1, data.UserID1, `{"exist":true}`)
	query.TestGetIdpNodes(t, 1, `{"node":[{"node_id":"`+data.IdP1+`","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"agent":false,"use_whitelist":false,"supported_request_message_data_url_type_list":[]}]}`)
	query.TestGetIdpNodesInfo(t, 1, `{"node":[{"node_id":"`+data.IdP1+`","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"agent":false,"use_whitelist":false,"supported_request_message_data_url_type_list":[]}]}`)
	query.TestGetIdpNodes(t, 2, `{"node":[]}`)
	query.TestGetIdpNodes(t, 3, `{"node":[{"node_id":"`+data.IdP1+`","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"agent":false,"use_whitelist":false,"ial":3,"mode_list":[2],"supported_request_message_data_url_type_list":[]}]}`)
	query.TestGetIdpNodes(t, 4, `{"node":[{"node_id":"`+data.IdP1+`","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"agent":false,"use_whitelist":false,"ial":3,"mode_list":[2],"supported_request_message_data_url_type_list":[]}]}`)
	query.TestGetIdpNodesInfo(t, 2, `{"node":[{"node_id":"`+data.IdP1+`","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"agent":false,"use_whitelist":false,"ial":3,"mode_list":[2],"supported_request_message_data_url_type_list":[]}]}`)

}

func TestIdP2RegisterIdentity(t *testing.T) {
	common.TestCreateRequest(t, data.RequestID2.String())
	idp.TestCreateIdpResponse(t, data.RequestID2.String())
	common.TestCloseRequest(t, data.RequestID2.String())
	idp.TestRegisterIdentity(t, 5, "Identity already existed")
	idp.TestRegisterIdentity(t, 6, "Identifier count is greater than allowed identifier count")
	idp.TestRegisterIdentity(t, 7, "There are duplicate identifier")
	idp.TestRegisterIdentity(t, 8, "success")
	query.TestGetIdpNodes(t, 5, `{"node":[{"node_id":"`+data.IdP1+`","node_name":"IdP Number 1","max_ial":3,"max_aal":3,"agent":false,"use_whitelist":false,"ial":3,"mode_list":[2],"supported_request_message_data_url_type_list":[]},{"node_id":"`+data.IdP2+`","node_name":"IdP Number 2","max_ial":2.3,"max_aal":3,"agent":false,"use_whitelist":false,"ial":2.3,"mode_list":[2],"supported_request_message_data_url_type_list":[]}]}`)
	query.TestGetReferenceGroupCodeByAccessorID(t, data.AccessorID3.String(), `{"reference_group_code":""}`)
	query.TestGetReferenceGroupCodeByAccessorID(t, data.AccessorID1.String(), `{"reference_group_code":"`+data.ReferenceGroupCode1.String()+`"}`)

}

func TestIdPAgent1GetNodesInfo(t *testing.T) {
	query.TestGetNodeInfo(t, data.IdPAgent1, `{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAz7N55vytQuBV17KHPzd1\nILPonOpltFqcMCV+x81NJNcvf2QcDDemYK2oObcs8rDuavx3+aSAeBrGXmFIjvVT\n7YTpEfoCGVf50AJKeyOeuaGefVy12GlGUsxKxCWDJaWe6Vc7S+cOyiLHNp/U/La3\nrSRbJeS6+GLbbVtJZpXsJwIejrK52JwSnCTH9aeVUDovJZNfQvPHaKArqermyI7/\n44o8qfGkImAs4UhLLpcQVyyADaqHMFKpRTE/cLISCB6Ut9Vb1lyBgk0xlGWLfrXa\n0erk96NK3tw0thd464qz2qFojNISmM1ddG+VSHoZUu7UJzeCUXyw0RkB1PZEXiwz\n7wIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Agent 1","role":"IdP","max_ial":2.3,"max_aal":3,"supported_request_message_data_url_type_list":[],"agent":true,"use_whitelist":false,"mq":[],"active":true}`)
}

func TestIdP2AddAccessor(t *testing.T) {
	common.TestCreateRequest(t, data.RequestID3.String())
	idp.TestCreateIdpResponse(t, data.RequestID3.String())
	common.TestCloseRequest(t, data.RequestID3.String())
	idp.TestAddAccessor(t, 1, "Found reference group code and identity detail in parameter")
	idp.TestAddAccessor(t, 2, "Reference group not found")
	idp.TestAddAccessor(t, 3, "success")
	query.TestGetReferenceGroupCodeByAccessorID(t, data.AccessorID3.String(), `{"reference_group_code":"`+data.ReferenceGroupCode1.String()+`"}`)
	query.TestGetReferenceGroupCode(t, 1, `{"reference_group_code":"`+data.ReferenceGroupCode1.String()+`"}`)
}

func TestNDIDAddService(t *testing.T) {
	ndid.TestAddService(t, data.ServiceID1)
}

func TestNDIDRegisterServiceDestinationByNDID(t *testing.T) {
	ndid.TestRegisterServiceDestinationByNDID(t, 1, "Node ID not found")
	ndid.TestRegisterServiceDestinationByNDID(t, 2, "Role of node ID is not AS")
	ndid.TestRegisterServiceDestinationByNDID(t, 3, "success")
	ndid.TestRegisterServiceDestinationByNDID(t, 4, "success")
}

func TestASRegisterServiceDestination(t *testing.T) {
	as.TestRegisterServiceDestination(t, 1, "success")
	query.TestGetAsNodesByServiceId(t, 1, `{"node":[{"node_id":"`+data.AS1+`","node_name":"AS1","min_ial":1.2,"min_aal":1.1,"supported_namespace_list":["`+data.UserNamespace1+`"]}]}`)
	query.TestGetAsNodesInfoByServiceId(t, 1, `{"node":[{"node_id":"`+data.AS1+`","name":"AS1","min_ial":1.2,"min_aal":1.1,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.102","port":8000}],"supported_namespace_list":["`+data.UserNamespace1+`"]}]}`)
	as.TestUpdateServiceDestination(t, 1, "success")
	query.TestGetAsNodesByServiceId(t, 1, `{"node":[{"node_id":"`+data.AS1+`","node_name":"AS1","min_ial":1.5,"min_aal":1.4,"supported_namespace_list":["`+data.UserNamespace2+`"]}]}`)
	query.TestGetAsNodesInfoByServiceId(t, 1, `{"node":[{"node_id":"`+data.AS1+`","name":"AS1","min_ial":1.5,"min_aal":1.4,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.102","port":8000}],"supported_namespace_list":["`+data.UserNamespace2+`"]}]}`)
	query.TestGetServicesByAsID(t, 1, `{"services":[{"service_id":"`+data.ServiceID1+`","min_ial":1.5,"min_aal":1.4,"active":true,"suspended":false,"supported_namespace_list":["`+data.UserNamespace2+`"]}]}`)
	as.TestRegisterServiceDestination(t, 2, "success")
	query.TestGetAsNodesByServiceId(t, 1, `{"node":[{"node_id":"`+data.AS1+`","node_name":"AS1","min_ial":1.5,"min_aal":1.4,"supported_namespace_list":["`+data.UserNamespace2+`"]},{"node_id":"`+data.AS2+`","node_name":"AS2","min_ial":1.2,"min_aal":1.1,"supported_namespace_list":["`+data.UserNamespace1+`"]}]}`)
	query.TestGetAsNodesInfoByServiceId(t, 1, `{"node":[{"node_id":"`+data.AS1+`","name":"AS1","min_ial":1.5,"min_aal":1.4,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApT8lXT9CDRZZkvhZLBD6\n6o7igZf6sj/o0XooaTuy2HuCt6yEO8jt7nx0XkEFyx4bH4/tZNsKdok7DU75MjqQ\nrdqGwpogvkZ3uUahwE9ZgOj6h4fq9l1Au8lxvAIp+b2BDRxttbHp9Ls9nK47B3Zu\niD02QknUNiPFvf+BWIoC8oe6AbyctnV+GTsC/H3jY3BD9ox2XKSE4/xaDMgC+SBU\n3pqukT35tgOcvcSAMVJJ06B3uyk19MzK3MVMm8b4sHFQ76UEpDOtQZrmKR1PH0gV\nFt93/0FPOH3m4o+9+1OStP51Un4oH3o80aw5g0EJzDpuv/+Sheec4+0PVTq0K6kj\ndQIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.102","port":8000}],"supported_namespace_list":["`+data.UserNamespace2+`"]},{"node_id":"`+data.AS2+`","name":"AS2","min_ial":1.2,"min_aal":1.1,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzhJ5PP3dfQtpw9p0Kphb\n30gg9jpgsv425D5pzZaH00zPgYfNTVZWfrLlTtc/ja8dbHvyDaCyzFD++Vr1vtmS\nSs9/j8ZhTJrTYHoiHvfG1ulTl1QdgwOcrKhpfhhjnCVCPOYjptgac/KPjhT7uiuY\nwB6axafx+RqPQqwQQhmuuxmTyy69l/cqezDtYCYUJVA6nV29ZaaF1VjWoE05PK16\n8mcB5quBdE6Vkc4n2k0wxaaTd/s9LPy6STXtz5IBXH2Gy5RP0TGeXO6iur/ZSM2z\n/3vQkTMjY/mkDduGioXcB6ieNgVv3XYbZg4VJEDSuOpRZReKcgLXvwk3CqZZdZRR\njQIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.103","port":8000}],"supported_namespace_list":["`+data.UserNamespace1+`"]}]}`)
}

func TestIdP1UpdateIdentity(t *testing.T) {
	query.TestGetIdentityInfo(t, 2, `{"ial":3,"mode_list":[2]}`)
	idp.TestUpdateIdentity(t, 1, "success")
	query.TestGetIdentityInfo(t, 2, `{"ial":2.3,"mode_list":[2]}`)
	query.TestGetIdpNodesInfo(t, 3, `{"node":[{"node_id":"`+data.IdP1+`","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"agent":false,"use_whitelist":false,"ial":2.3,"mode_list":[2],"supported_request_message_data_url_type_list":[]},{"node_id":"`+data.IdP2+`","name":"IdP Number 2","max_ial":2.3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.100","port":8000}],"agent":false,"use_whitelist":false,"ial":2.3,"mode_list":[2],"supported_request_message_data_url_type_list":[]}]}`)
}

func TestIdP2RevokeIdentityAssociation(t *testing.T) {
	common.TestCreateRequest(t, data.RequestID4.String())
	idp.TestCreateIdpResponse(t, data.RequestID4.String())
	common.TestCloseRequest(t, data.RequestID4.String())
	idp.TestRevokeIdentityAssociation(t, 1, "success")
	query.TestGetIdpNodesInfo(t, 3, `{"node":[{"node_id":"`+data.IdP1+`","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"agent":false,"use_whitelist":false,"ial":2.3,"mode_list":[2],"supported_request_message_data_url_type_list":[]}]}`)
}

func TestIdP2RegisterIdentityAfterRevokeIdentityAssociation(t *testing.T) {
	common.TestCreateRequest(t, data.RequestID5.String())
	idp.TestCreateIdpResponse(t, data.RequestID5.String())
	common.TestCloseRequest(t, data.RequestID5.String())
	idp.TestRegisterIdentity(t, 9, "success")
	query.TestGetIdpNodesInfo(t, 3, `{"node":[{"node_id":"`+data.IdP1+`","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"agent":false,"use_whitelist":false,"ial":2.3,"mode_list":[2],"supported_request_message_data_url_type_list":[]},{"node_id":"`+data.IdP2+`","name":"IdP Number 2","max_ial":2.3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.100","port":8000}],"agent":false,"use_whitelist":false,"ial":2.3,"mode_list":[2,3],"supported_request_message_data_url_type_list":[]}]}`)
}

func TestQueryGetAccessorKey(t *testing.T) {
	query.TestGetAccessorKey(t, data.AccessorID1.String(), `{"accessor_public_key":"`+strings.Replace(data.AccessorPubKey1, "\n", "\\n", -1)+`","active":true}`)
	query.TestGetAccessorKey(t, data.AccessorID2.String(), `{"accessor_public_key":"`+strings.Replace(data.AccessorPubKey2, "\n", "\\n", -1)+`","active":true}`)
}
func TestQueryGetAllowedModeList(t *testing.T) {
	query.TestGetAllowedModeList(t, "", `{"allowed_mode_list":[1,2,3]}`)
}

func TestIdP1UpdateIdentityModeList(t *testing.T) {
	query.TestGetIdentityInfo(t, 2, `{"ial":2.3,"mode_list":[2]}`)
	idp.TestUpdateIdentityModeList(t, 1, "success")
	query.TestGetIdentityInfo(t, 2, `{"ial":2.3,"mode_list":[2,3]}`)
}

func TestIdP1AddIdentity(t *testing.T) {
	common.TestCreateRequest(t, data.RequestID6.String())
	idp.TestCreateIdpResponse(t, data.RequestID6.String())
	common.TestCloseRequest(t, data.RequestID6.String())
	idp.TestAddIdentity(t, 1, "success")
	query.TestGetIdentityInfo(t, 3, `{"ial":2.3,"mode_list":[2,3]}`)
}

func TestIdP1UpdateNode(t *testing.T) {
	query.TestGetNodeInfo(t, data.IdP1, `{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 1","role":"IdP","max_ial":3,"max_aal":3,"supported_request_message_data_url_type_list":[],"agent":false,"use_whitelist":false,"mq":[{"ip":"192.168.3.99","port":8000}],"active":true}`)
	common.TestUpdateNode(t, 1, "success")
	query.TestGetNodeInfo(t, data.IdP1, `{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 1","role":"IdP","max_ial":3,"max_aal":3,"supported_request_message_data_url_type_list":["text/plain","application/pdf"],"agent":false,"use_whitelist":false,"mq":[{"ip":"192.168.3.99","port":8000}],"active":true}`)
	query.TestGetIdpNodesInfo(t, 4, `{"node":[{"node_id":"`+data.IdP1+`","name":"IdP Number 1","max_ial":3,"max_aal":3,"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEArdcKj/gAetVyg6Nn2lDi\nm/UJYQsQCav60EVbECm5EVT8WgnpzO+GrRyBtxqWUdtGar7d6orLh1RX1ikU7Yx2\nSA8Xlf+ZDaCELba/85Nb+IppLBdPywixgumoto9G9dDGSnPkHAlq5lXXA1eeUS7j\niU1lf37lwTZaO0COAuu8Vt9GcwYPh7SSf4/eXabQGbo/TMUVpXX1w5N1A07Qh5DG\nr/ZKzEE9/5bJJJRS635OA2T4gIY9XRWYiTxtiZz6AFCxP92Cjz/sNvSc/Cuvwi15\nycS4C35tjM8iT5djsRcR+MJeXyvurkaYgMGJTDIWub/A5oavVD3VwusZZNZvpDpD\nPwIDAQAB\n-----END PUBLIC KEY-----\n","mq":[{"ip":"192.168.3.99","port":8000}],"agent":false,"use_whitelist":false,"ial":2.3,"mode_list":[2,3],"supported_request_message_data_url_type_list":["text/plain","application/pdf"]}]}`)
	common.TestUpdateNode(t, 2, "success")
	query.TestGetNodeInfo(t, data.IdP1, `{"public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwx9oT44DmDRiQJ1K0b9Q\nolEsrQ51hBUDq3oCKTffBikYenSUQNimVCsVBfNpKhZqpW56hH0mtgLbI7QgZGj9\ncNBMzSLMolltw0EerF0Ckz0Svvie1/oFJ1a0Cf4bdKKW6wRzL+aFVvelmNlLoSZX\noCpxUPQq7SMLoYEK1c+e3l3H0bfh6TAVt7APOQEFhXy9MRt83oVSAGW36gdNEksm\nz1WIT/C1XcHHVwCIJGSdZw5F6Y2gBjtiLsiFtpKfxQAPwBvDi7uS0PUdN7YQ/G69\nb0FgoE6qivDTqYfr80Y345Qe/qPGDvfne7oA8DIbRV+Kd5s4tFn/cC0Wd+jvrZJ7\njwIDAQAB\n-----END PUBLIC KEY-----\n","master_public_key":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAukTxVg8qpwXebALGCrly\niv8PNNxLo0CEX3N33cR1TNfImItd5nFwmozLJLM9LpNF711PrkH3EBLJM+qwASlC\nBayeMiMT8tDmOtv1RqIxyLjEU8M0RBBedk/TsKQwNmmeU3n5Ap+GRTYoEOwTKNra\nI8YDfbjb9fNtSICiDzn3UcQj13iLz5x4MjaewtC6PR1r8uVfLyS4uI+3/qau0zWV\n+s6b3JdqU2zdHeuaj9XjX7aNV7mvnjYgzk/O7M/p/86RBEOm7pt6JmTGnFu44jBO\nez6GqF2hZzqR9nM1K4aOedBMHintVnhh1oOPG9uRiDnJWvN16PNTfr7XBOUzL03X\nDQIDAQAB\n-----END PUBLIC KEY-----\n","node_name":"IdP Number 1","role":"IdP","max_ial":3,"max_aal":3,"supported_request_message_data_url_type_list":["text/plain"],"agent":false,"use_whitelist":false,"mq":[{"ip":"192.168.3.99","port":8000}],"active":true}`)
}

func TestIdP1RevokeAndAddAccessor(t *testing.T) {
	query.TestGetAccessorKey(t, data.AccessorID1.String(), `{"accessor_public_key":"`+strings.Replace(data.AccessorPubKey1, "\n", "\\n", -1)+`","active":true}`)
	common.TestCreateRequest(t, data.RequestID7.String())
	idp.TestCreateIdpResponse(t, data.RequestID7.String())
	common.TestCloseRequest(t, data.RequestID7.String())
	idp.TestRevokeAndAddAccessor(t, 1, "success")
	query.TestGetAccessorKey(t, data.AccessorID1.String(), `{"accessor_public_key":"`+strings.Replace(data.AccessorPubKey1, "\n", "\\n", -1)+`","active":false}`)
	query.TestGetAccessorKey(t, data.AccessorID5.String(), `{"accessor_public_key":"`+strings.Replace(data.AccessorPubKey2, "\n", "\\n", -1)+`","active":true}`)

}

func TestAddErrorCodeByNDID(t *testing.T) {
	ndid.TestAddErrorCode(t, "idp", data.IdpErrorCode1, data.IdpErrorCodeDescription1, false, "success")
	ndid.TestAddErrorCode(t, "idp", data.IdpErrorCode1, data.IdpErrorCodeDescription1, false, "ErrorCode is already in the database")
	query.TestGetErrorCodeList(t, "idp", `[{"error_code":"`+data.IdpErrorCode1+`","description":"`+data.IdpErrorCodeDescription1+`","fatal":false}]`)
	query.TestGetErrorCodeList(t, "as", "[]")
}

func TestRemoveErrorCodeByNDID(t *testing.T) {
	ndid.TestRemoveErrorCode(t, "idp", data.IdpErrorCode1, "success")
	ndid.TestRemoveErrorCode(t, "idp", data.IdpErrorCode1, "ErrorCode not exists")
	query.TestGetErrorCodeList(t, "idp", `[]`)
}
